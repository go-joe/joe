+++
title = "Handling Custom Events"
slug = "events"
weight = 3
pre = "<b>c) </b>"
+++

{{% notice note %}}
More information on the events system can now be found in the [**Events recipe**](/recipes/events)
{{% /notice %}}

The previous example should give you an idea already on how to write simple chat
bots. It is missing one important part however: how can a bot trigger any
interaction proactively, i.e. without a message from a user.

To solve this problem, joe's Brain implements an event handler that you can hook
into. In fact the `Bot.Respond(…)` function that we used in the earlier examples
is doing exactly that to listen for any `joe.ReceiveMessageEvent` that match the
specified regular expression and then execute the handler function.

Implementing custom events is easy because you can emit any type as event and
register handlers that match only this type. What this exactly means is best
demonstrated with another example:

[embedmd]:# (_examples/03_custom_events/main.go)
```go
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
```
