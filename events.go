package joe

type InitEvent struct{}

type ShutdownEvent struct{}

type ReceiveMessageEvent struct {
	Text      string
	ChannelID string
}
