+++
title = "Reactions & Emojis"
slug = "reactions"
weight = 6
pre = "<b>f) </b>"
+++

Joe supports reacting to messages with Emojis if the chat adapter has this feature.
For instance in Slack your bot could add a :robot: emoji to a message to indicate to
the user, that the bot has seen it. Another use case of reactions is that you
may want to trigger an action if the _user_ attaches an emoji to a message.

Currently the following chat adapters support reactions:

- <i class="fas fa-terminal"></i> CLI Adapter: https://github.com/go-joe/joe
- <i class='fab fa-slack fa-fw'></i> Slack Adapter: https://github.com/go-joe/slack-adapter
- <i class='fab fa-rocketchat fa-fw'></i> Rocket.Chat Adapter: https://github.com/dwmunster/rocket-adapter

The following example shows how you can use reactions in your message handlers:

[embedmd]:# (_examples/06_react/main.go)
```go
package main

import (
	"fmt"

	"github.com/go-joe/joe"
	"github.com/go-joe/joe/reactions"
)

func main() {
	b := joe.New("example-bot")
	b.Respond("hello", MyHandler)
	b.Brain.RegisterHandler(ReceiveReaction)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func MyHandler(msg joe.Message) error {
	err := msg.React(reactions.Thumbsup)
	if err != nil {
		msg.Respond("Sorry but there was an issue attaching a reaction: %v", err)
	}

	// custom reactions are also possible
	_ = msg.React(reactions.Reaction{Shortcode: "foo"})

	return err
}

func ReceiveReaction(evt reactions.Event) error {
	fmt.Printf("Received event: %+v", evt)
	return nil
}
```

If you try to react to a message, when your chat adapter does not support this
feature, the `Message.React(…)` function will return the `joe.ErrNotImplemented`
sentinel error.
