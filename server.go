package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	return slackutil.Attachment{
		Text: fmt.Sprintf(
			"Instance <%s|%s> is a `%s` `%s` in `%s`",
			instance.GetLink("ec2_console"),
			instance.GetMetadata("instance_id"),
			instance.GetMetadata("instance_state"),
			instance.GetMetadata("instance_type"),
			instance.GetMetadata("az"),
		),
		Fields: []slackutil.Field{
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
			slackutil.Field{
				Title: "Public IP(s)",
				Value: instance.GetMetadata("public_ips"),
				Short: true,
			},
			slackutil.Field{
				Title: "Private IP(s)",
				Value: instance.GetMetadata("private_ips"),
				Short: true,
			},
			slackutil.Field{
				Value: fmt.Sprintf("⏳ <%s|AWS config timeline>", instance.GetLink("config_timeline")),
			},
		},
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
			Text: "One second while we look that up...",
		},

		Handler: func(ctx context.Context, req slackutil.SlashCommandRequest, resp slackutil.MessageResponder) {
			time.Sleep(time.Second)
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
	}

	findResources.Run(w, *command)
}
