package main

import (
	"time"

	"github.com/go-joe/joe"
)

type ExampleBot struct {
	*joe.Bot
	Channel string // example for your custom bot configuration
}

type CustomEvent struct {
	Data string // just an example of attaching any data with a custom event
}

func main() {
	b := &ExampleBot{
		Bot:     joe.New("example"),
		Channel: "CDEADBEAF", // example reference to a slack channel
	}

	// Register our custom event handler. Joe inspects the function signature to
	// understand that this function should be invoked whenever a CustomEvent
	// is emitted.
	b.Brain.RegisterHandler(b.HandleCustomEvent)

	// For example purposes emit a CustomEvent in a second.
	time.AfterFunc(time.Second, func() {
		b.Brain.Emit(CustomEvent{Data: "Hello World!"})
	})

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

// HandleCustomEvent handles any CustomEvent that is emitted. Joe also supports
// event handlers that return an error or accept a context.Context as first argument.
func (b *ExampleBot) HandleCustomEvent(evt CustomEvent) {
	b.Say(b.Channel, "Received custom event: %v", evt.Data)
}
