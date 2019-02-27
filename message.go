package joe

import (
	"context"
	"fmt"
)

// A Message is automatically created from a ReceiveMessageEvent and then passed
// to the RespondFunc that was passed to Bot.Respond(…) or Bot.RespondRegex(…)
// when the message matches the regular expression of the handler.
type Message struct {
	Context context.Context
	Text    string
	Channel string
	Matches []string // contains all sub matches of the regex

	adapter Adapter
}

func (msg *Message) Respond(text string, args ...interface{}) {
	_ = msg.RespondE(text, args...)
}

func (msg *Message) RespondE(text string, args ...interface{}) error {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}

	return msg.adapter.Send(text, msg.Channel)
}
