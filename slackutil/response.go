package slackutil

import (
	"encoding/json"
	"net/http"
	"time"
)

// Copied from https://github.com/signalsciences/slashcmd/blob/master/response.go
const (
	// The response should be visible to everyone in the channel
	ResponseInChannel = "in_channel"
	// The response should only be visible to the user who typed the command
	ResponseEphemeral = "ephemeral"
)

// Response is the response to a slash command
type Response struct {
	ResponseType string       `json:"response_type,omitempty"`
	Text         string       `json:"text"`
	Attachments  []Attachment `json:"attachments,omitempty"`
}

// Attachment is Slack attachment for slash Response
type Attachment struct {
	Fallback      string   `json:"fallback"`
	Text          string   `json:"text"`
	MarkdownIn    []string `json:"mrkdwn_in,omitempty"`
	Color         string   `json:"color,omitempty"`
	AuthorName    string   `json:"author_name,omitempty"`
	AuthorSubname string   `json:"author_subname,omitempty"`
	AuthorLink    string   `json:"author_link,omitempty"`
	AuthorIcon    string   `json:"author_icon,omitempty"`
	Title         string   `json:"title,omitempty"`
	TitleLink     string   `json:"title_link,omitempty"`
	Pretext       string   `json:"pretext,omitempty"`
	ImageURL      string   `json:"image_url,omitempty"`
	ThumbURL      string   `json:"thumb_url,omitempty"`
	Fields        []Field  `json:"fields,omitempty"`
	Footer        string   `json:"footer,omitempty"`
	FooterIcon    string   `json:"footer_icon,omitempty"`
	Timestamp     int64    `json:"ts,omitempty"`
}

// Field is a field attachment
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Timestamp formats a time.Time into a unix epoch
// value suitable for Attachment.Timestamp
//
// Timestamp: debatable if this should be included
//
func Timestamp(t time.Time) int64 {
	return t.UTC().Unix()
}

func RespondWith(w http.ResponseWriter, resp Response) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
