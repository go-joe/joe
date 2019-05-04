package joe

import (
	"fmt"

	"github.com/pkg/errors"
)

// The InitEvent is the first event that is handled by the Brain after the Bot
// is started via Bot.Run().
type InitEvent struct{}

// The RegisterCommandEvent is emitted when a new message handler is registered
// via Bot.Respond(…) or Bot.RespondRegex(…).
type RegisterCommandEvent struct {
	Expression string
	Function   func(Message) error
}

// The ShutdownEvent is the last event that is handled by the Brain before it
// stops handling any events after the bot context is done.
type ShutdownEvent struct{}

// The ReceiveMessageEvent is typically emitted by an Adapter when the Bot sees
// a new message from the chat.
type ReceiveMessageEvent struct {
	ID       string // The ID of the message, identifying it at least uniquely within the Channel
	Text     string // The message text.
	AuthorID string // A string identifying the author of the message on the adapter.
	Channel  string // The channel over which the message was received.
	Adapter Adapter // The adapter that has emitted this event

	// A message may optionally also contain additional information that was
	// received by the Adapter (e.g. with the slack adapter this may be the
	// *slack.MessageEvent. Each Adapter implementation should document if and
	// what information is available here, if any at all.
	Data interface{}
}

// The UserTypingEvent is emitted by the Adapter and indicates that the Bot
// sees that a user is typing. This event may not be emitted on all Adapter
// implementations but only when it is actually supported (e.g. on slack).
type UserTypingEvent struct {
	User    User
	Channel string
}

// Respond is a helper function to directly send a response back to the channel
// the message originated from. This function ignores any error when sending the
// response. If you want to handle the error use Message.RespondE instead.
func (evt *ReceiveMessageEvent) Respond(text string, args ...interface{}) {
	_ = evt.RespondE(text, args...)
}

// RespondE is a helper function to directly send a response back to the channel
// the message originated from. If there was an error it will be returned from
// this function.
func (evt *ReceiveMessageEvent) RespondE(text string, args ...interface{}) error {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}

	if evt.Adapter == nil {
		return errors.New("misbehaving adapter: the Adapter field of the ReceiveMessageEvent is nil")
	}

	return evt.Adapter.Send(text, evt.Channel)
}
