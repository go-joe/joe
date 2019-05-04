package joe

import (
	"context"

	"github.com/go-joe/joe/reactions"
)

// A Message is automatically created from a ReceiveMessageEvent and then passed
// to the RespondFunc that was registered via Bot.Respond(…) or Bot.RespondRegex(…)
// when the message matches the regular expression of the handler.
type Message struct {
	ReceiveMessageEvent
	Context context.Context
	Matches []string // contains all sub matches of the regular expression that matched the Text
}

// React attempts to let the Adapter attach the given reaction to this message.
// If the adapter does not support this feature this function will return
// ErrNotImplemented.
func (msg *Message) React(reaction reactions.Reaction) error {
	adapter, ok := msg.Adapter.(ReactionAwareAdapter)
	if !ok {
		return ErrNotImplemented
	}

	return adapter.React(reaction, *msg)
}
