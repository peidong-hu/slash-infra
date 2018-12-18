package slackutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	bugsnag "github.com/bugsnag/bugsnag-go"
)

var slackClient = http.Client{}

type DelayedSlashResponse struct {
	// A mesage to send the user while we're preparing a response to
	PendingResponse Response

	Handler func(context.Context, SlashCommandRequest, MessageResponder)
}

func (d DelayedSlashResponse) Run(w http.ResponseWriter, command SlashCommandRequest) {
	RespondWith(w, d.PendingResponse)

	ctx := context.Background()

	responder := MessageResponder{command}

	// Not using a waitgroup here as we don't really care about cleaning up this goroutine
	go func() {
		defer bugsnag.AutoNotify(ctx)
		d.Handler(ctx, command, responder)
	}()
}

type MessageResponder struct {
	command SlashCommandRequest
}

func (m MessageResponder) PublicResponse(resp Response) {
	resp.ResponseType = ResponseInChannel

	b, err := json.Marshal(&resp)
	if err != nil {
		panic(err)
	}

	r, err := http.NewRequest("POST", m.command.ResponseURL, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	apiResp, err := slackClient.Do(r)
	if err != nil {
		panic(err)
	}

	fmt.Println(apiResp)
}

type SlashCommandResponder interface {
	PublicResponse(Response)
}
