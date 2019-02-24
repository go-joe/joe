package main

import (
	"log"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/fgrosse/joe"
	"github.com/fgrosse/joe/redis-brain"
	"github.com/fgrosse/joe/slack-adapter"
)

type ExampleBot struct {
	*joe.Bot
}

// IDEA:
// - generate code based on godoc of functions?
// - lint code (e.g. using match indices that cannot exist
// - collect all settings in a concrete type and implement reading from env or file there (viper?)

func main() {
	b := &ExampleBot{Bot: joe.New("example",
		redis.Brain("localhost:6379", redis.WithKey("joe")),
		slack.Adapter("xoxb-17858453111-558911412836-sOj22lLot5qSLXfVnLD6UKE4"),
	)}

	b.Respond("ping", b.Pong)
	b.Respond("remember (.+) is (.+)", b.Remember)
	b.Respond(`(what is|remember) ([^?]+)\s*\??`, b.WhatIs)
	b.Respond(`forget (.+)`, b.Forget)
	b.Respond(`what do you remember\??`, b.WhatDoYouRemember)

	err := b.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func (b *ExampleBot) Pong(msg joe.Message) error {
	msg.Respond("PONG")
	return nil
}

// Remember a value for a given key.
//   command: bot remember <key> is <value>
func (b *ExampleBot) Remember(msg joe.Message) error {
	key, value := msg.Matches[0], msg.Matches[1]
	key = strings.TrimSpace(key)
	msg.Respond("OK, I'll remember %s is %s", key, value)
	return b.Brain.Set(key, value)
}

func (b *ExampleBot) WhatIs(msg joe.Message) error {
	key := strings.TrimSpace(msg.Matches[1])
	value, ok, err := b.Brain.Get(key)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve key %q from brain", key)
	}

	if ok {
		msg.Respond("%s is %s", key, value)
	} else {
		msg.Respond("I do not remember %q", key)
	}

	return nil
}

func (b *ExampleBot) Forget(msg joe.Message) error {
	key := strings.TrimSpace(msg.Matches[0])
	value, _, _ := b.Brain.Get(key)
	ok, err := b.Brain.Delete(key)
	if err != nil {
		return errors.Wrapf(err, "failed to delete key %q from brain", key)
	}

	if !ok {
		msg.Respond("I do not remember %q", key)
	} else {
		msg.Respond("I've forgotten %s is %s.", key, value)
	}

	return nil
}

func (b *ExampleBot) WhatDoYouRemember(msg joe.Message) error {
	data, err := b.Brain.Memories()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve all memories from brain")
	}

	switch len(data) {
	case 0:
		msg.Respond("I do not remember anything")
		return nil
	case 1:
		msg.Respond("I have only a single memory:")
	default:
		msg.Respond("I have %d memories:", len(data))
	}

	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for _, key := range keys {
		value := data[key]
		msg.Respond("%s is %s", key, value)
	}

	return nil
}
