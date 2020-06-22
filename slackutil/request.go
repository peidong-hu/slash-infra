package slackutil

import "net/http"

// SlashCommandRequest is a payload that slack sends your app when a user
// invokes a slash command
type SlashCommandRequest struct {
	// The command that was typed in to trigger this request. This value
	// can be useful if you want to use a single Request URL to service
	// multiple Slash Commands, as it lets you tell them apart.
	Command string

	// These IDs provide context about where the user was in Slack when
	// they triggered your app's command (eg. which workspace, Enterprise
	// Grid, or channel). You may need these IDs for your command response.
	TeamID         string
	TeamDomain     string
	EnterpriseID   string
	EnterpriseName string
	ChannelID      string
	ChannelName    string

	// The slack user that invoked this slash command
	UserID   string
	UserName string

	// This is the part of the Slash Command after the command itself, and
	// it can contain absolutely anything that the user might decide to
	// type. It is common to use this text parameter to provide extra
	// context for the command.
	Text string

	// A URL that you can use to respond to the command.
	ResponseURL string

	// If you need to respond to the command by opening a dialog, you'll
	// need this trigger ID to get it to work. You can use this ID with
	// dialog.open up to 3000ms after this data payload is sent.
	TriggerID string
}

func ParseSlashCommandRequest(r *http.Request) (*SlashCommandRequest, error) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}
	var query *url.Values
	if len(input) > 0 {
		q, err := url.ParseQuery(string(input))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query = &q
	}
	return &SlashCommandRequest{
		Command:        query.Get("command"),
		TeamID:         query.Get("team_id"),
		TeamDomain:     query.Get("team_domain"),
		EnterpriseID:   query.Get("enterprise_id"),
		EnterpriseName: query.Get("enterprise_name"),
		ChannelID:      query.Get("channel_id"),
		ChannelName:    query.Get("channel_name"),
		UserID:         query.Get("user_id"),
		UserName:       query.Get("user_name"),
		Text:           query.Get("text"),
		ResponseURL:    query.Get("response_url"),
		TriggerID:      query.Get("trigger_id"),
	}, nil
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	return &SlashCommandRequest{
		Command:        r.PostForm.Get("command"),
		TeamID:         r.PostForm.Get("team_id"),
		TeamDomain:     r.PostForm.Get("team_domain"),
		EnterpriseID:   r.PostForm.Get("enterprise_id"),
		EnterpriseName: r.PostForm.Get("enterprise_name"),
		ChannelID:      r.PostForm.Get("channel_id"),
		ChannelName:    r.PostForm.Get("channel_name"),
		UserID:         r.PostForm.Get("user_id"),
		UserName:       r.PostForm.Get("user_name"),
		Text:           r.PostForm.Get("text"),
		ResponseURL:    r.PostForm.Get("response_url"),
		TriggerID:      r.PostForm.Get("trigger_id"),
	}, nil
}
