package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/geckoboard/slash-infra/search"
	"github.com/geckoboard/slash-infra/slackutil"
	"github.com/julienschmidt/httprouter"
)

func makeHttpHandler() *httprouter.Router {
	router := httprouter.New()

	s := httpServer{
		ec2Resolver: search.NewEc2(),
	}

	router.POST("/slack/infra-search", s.whatIsHandler)

	return router
}

type httpServer struct {
	ec2Resolver *search.EC2Resolver
}

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(msg))
}

func FormatEc2InstanceAsAttachment(instance search.Result) slackutil.Attachment {
	fields := []slackutil.Field{
		slackutil.Field{
			Title: "Environment",
			Value: instance.GetMetadata("tag:Environment"),
			Short: true,
		},
		slackutil.Field{
			Title: "Role",
			Value: instance.GetMetadata("tag:Role"),
			Short: true,
		},
	}

	if publicIps := instance.GetMetadata("public_ips"); publicIps != "" {
		fields = append(fields, slackutil.Field{
			Title: "Public IP(s)",
			Value: instance.GetMetadata("public_ips"),
			Short: true,
		})
	}
	if privateIps := instance.GetMetadata("private_ips"); privateIps != "" {
		fields = append(fields, slackutil.Field{
			Title: "Private IP(s)",
			Value: privateIps,
			Short: true,
		})
	}
	fields = append(fields, slackutil.Field{
		Value: fmt.Sprintf("⏳ <%s|AWS config timeline>", instance.GetLink("config_timeline")),
	})

	return slackutil.Attachment{
		Text: fmt.Sprintf(
			"Instance <%s|%s> is a `%s` `%s` in `%s`",
			instance.GetLink("ec2_console"),
			instance.GetMetadata("instance_id"),
			instance.GetMetadata("instance_state"),
			instance.GetMetadata("instance_type"),
			instance.GetMetadata("az"),
		),
		Fields:     fields,
		MarkdownIn: []string{"text"},
	}
}

func (h httpServer) whatIsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	command, err := slackutil.ParseSlashCommandRequest(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not parse payload")
		return
	}

	findResources := slackutil.DelayedSlashResponse{
		PendingResponse: slackutil.Response{
			Text: "Hang on a jiffy while we look that up...",
		},

		Handler: func(ctx context.Context, req slackutil.SlashCommandRequest, resp slackutil.MessageResponder) {
			resultSets := h.ec2Resolver.Search(ctx, command.Text)

			response := slackutil.Response{
				Attachments: []slackutil.Attachment{},
			}

			for _, setOfResults := range resultSets {
				if setOfResults.Kind == "ec2.instance" {
					if len(setOfResults.Results) == 1 {
						response.Attachments = append(response.Attachments, FormatEc2InstanceAsAttachment(setOfResults.Results[0]))
					}
				}

			}

			resp.PublicResponse(response)

		},

		ShowSlashCommandInChannel: true,
	}

	findResources.Run(w, *command)
}
