+++
title = "Useful Example"
slug = "useful"
weight = 2
pre = "<b>b) </b>"
+++

Each bot consists of a chat _Adapter_ (e.g. to integrate with Slack), a _Memory_
implementation to remember key-value data (e.g. using Redis) and a _Brain_ which
routes new messages or custom events (e.g. receiving an HTTP call) to the
corresponding registered _handler_ functions.

By default `joe.New(…)` uses the CLI adapter which makes the bot read messages
from stdin and respond on stdout. Additionally the bot will store key value
data in-memory which means it will forget anything you told it when it is restarted.
This default setup is useful for local development without any dependencies but
you will quickly want to add other [_Modules_](/modules) to extend the bots capabilities.

For instance we can extend the previous example to connect the Bot with a Slack
workspace and store key-value data in Redis. To allow the message handlers to
access the memory we define them as functions on a custom `ExampleBot`type which
embeds the `joe.Bot`.

[embedmd]:# (../../../_examples/02_useful/main.go)
```go
package main

import (
	"fmt"

	"github.com/go-joe/joe"
	"github.com/go-joe/redis-memory"
	"github.com/go-joe/slack-adapter"
)

type ExampleBot struct {
	*joe.Bot
}

func main() {
	b := &ExampleBot{
		Bot: joe.New("example",
			redis.Memory("localhost:6379"),
			slack.Adapter("xoxb-1452345…"),
		),
	}

	b.Respond("remember (.+) is (.+)", b.Remember)
	b.Respond("what is (.+)", b.WhatIs)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func (b *ExampleBot) Remember(msg joe.Message) error {
	key, value := msg.Matches[0], msg.Matches[1]
	msg.Respond("OK, I'll remember %s is %s", key, value)
	return b.Store.Set(key, value)
}

func (b *ExampleBot) WhatIs(msg joe.Message) error {
	key := msg.Matches[0]
	var value string
	ok, err := b.Store.Get(key, &value)
	if err != nil {
		return fmt.Errorf("failed to retrieve key %q from brain: %w", key, err)
	}

	if ok {
		msg.Respond("%s is %s", key, value)
	} else {
		msg.Respond("I do not remember %q", key)
	}

	return nil
}
```
