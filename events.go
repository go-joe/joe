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

type BrainMemoryEvent struct {
	Key       string
	Value     string
	Operation string // "set", "get" or "del"
}
