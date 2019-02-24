package joe

type InitEvent struct{}

type ShutdownEvent struct{}

type ReceiveMessageEvent struct {
	Text    string
	Channel string
}

type UserTypingEvent struct {
	User    User
	Channel string
}
