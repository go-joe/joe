package joe

// The InitEvent is the first event that is handled by the Brain after the Bot
// is started via Bot.Run().
type InitEvent struct{}

// The ShutdownEvent is the last event that is handled by the Brain before it
// stops handling any events after the bot context is done.
type ShutdownEvent struct{}

// The ReceiveMessageEvent is typically emitted by an Adapter when the Bot sees
// a new message from the chat.
type ReceiveMessageEvent struct {
	Text     string // The message text.
	AuthorID string // A string identifying the author of the message on the adapter.
	Channel  string // The channel over which the message was received.

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
