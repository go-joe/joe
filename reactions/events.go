package reactions

// An Event may be emitted by a chat Adapter to indicate that a message
// received a reaction.
type Event struct {
	Reaction  Reaction
	MessageID string
	Channel   string
	AuthorID  string
}
