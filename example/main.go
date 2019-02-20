package main

import (
	"sort"
	"strings"

	"github.com/pkg/errors"

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
	b.Respond("remember (.+) is (.+)", b.Remember)
	b.Respond(`(what is|remember) ([^?]+)\s*\??`, b.WhatIs)
	b.Respond(`forget (.+)`, b.Forget)
	b.Respond(`what do you remember\??`, b.WhatDoYouRemember)

	b.Run()
}

func (b *ExampleBot) Pong(msg joe.Message) error {
	b.Say("PONG")
	return nil
}

// Remember a value for a given key.
//   command: bot remember <key> is <value>
func (b *ExampleBot) Remember(msg joe.Message) error {
	key, value := msg.Matches[0], msg.Matches[1]
	key = strings.TrimSpace(key)
	b.Say("OK, I'll remember %s is %s", key, value)
	return b.Brain.Set(key, value)
}

func (b *ExampleBot) WhatIs(msg joe.Message) error {
	key := strings.TrimSpace(msg.Matches[1])
	value, ok, err := b.Brain.Get(key)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve key %q from brain", key)
	}

	if ok {
		b.Say("%s is %s", key, value)
	} else {
		b.Say("I do not remember %q", key)
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
		b.Say("I do not remember %q", key)
	} else {
		b.Say("I've forgotten %s is %s.", key, value)
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
		b.Say("I do not remember anything")
		return nil
	case 1:
		b.Say("I have only a single memory:")
	default:
		b.Say("I have %d memories:", len(data))
	}

	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for _, key := range keys {
		value := data[key]
		b.Say("%s is %s", key, value)
	}

	return nil
}
