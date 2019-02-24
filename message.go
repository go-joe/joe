package joe

import (
	"context"
	"fmt"
)

type Message struct {
	Context   context.Context
	Text      string
	ChannelID string
	Matches   []string

	adapter Adapter
}

func (msg *Message) Respond(text string, args ...interface{}) {
	_ = msg.RespondE(text, args...)
}

func (msg *Message) RespondE(text string, args ...interface{}) error {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}

	return msg.adapter.Send(text, msg.ChannelID)
}
