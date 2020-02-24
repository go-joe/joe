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
