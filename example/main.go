package main

import (
	"go.uber.org/zap"

	"github.com/fgrosse/joe"
)

type ExampleBot struct {
	*joe.Bot
}

// IDEAS:
// - generate code based on godoc of functions?
// - lint code (e.g. using match indices that cannot exist
//

func main() {
	b := &ExampleBot{Bot: joe.New("example")}

	b.Respond("ping", b.Pong)
	b.Respond(`remember (.*) is (.*)`, b.Remember) // auto generated regular expressions
	// TODO: b.RespondRegex(`remember\s+(.*)\s+is\s+(.*)`, b.Remember) // full control over regular expression

	b.Run()
}

func (b *ExampleBot) Pong(msg joe.Message) error {
	b.Say("PONG")
	return nil
}

// Remember a value for a given key.
//   command: bot remember <key> is <value>
func (b *ExampleBot) Remember(msg joe.Message) error {
	b.Logger.Debug("Debug", zap.Strings("msg", msg.Matches))
	b.Say("OK, I'll remember %s is %s", msg.Matches[0], msg.Matches[1])
	return nil
}
