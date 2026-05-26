package delivery

import "fmt"

type TargetType string

const (
	TargetTypeNone        TargetType = ""
	TargetTypeMessage     TargetType = "message"
	TargetTypeInteraction TargetType = "interaction"
)

type Target struct {
	Type TargetType

	ChannelID string
	GuildID   string
	MessageID string

	AppID            string
	InteractionID    string
	InteractionToken string

	Filename             string
	FilenameBase         string
	ContentType          string
	FailureMessagePrefix string
}

func (t Target) IsZero() bool {
	return t.Type == TargetTypeNone
}

func (t Target) WithFormat(format string) Target {
	if t.Filename == "" && t.FilenameBase != "" {
		t.Filename = fmt.Sprintf("%s.%s", t.FilenameBase, format)
	}
	if t.ContentType == "" && format != "" {
		t.ContentType = "image/" + format
	}
	return t
}
