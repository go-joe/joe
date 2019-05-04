package main

import (
	"github.com/go-joe/joe"
	"github.com/go-joe/joe/help"
)

type Bot struct {
	*joe.Bot
}

func main() {
	b := Bot{joe.New("example-bot",
		help.Adapter(),
	)}

	b.Respond("ping", b.Pong)
	b.Respond("foo", b.Foo)
	b.Respond("hello", b.Hello)
	b.Respond("global", Global)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

// Pong does something useful.
func (*Bot) Pong(msg joe.Message) error {
	msg.Respond("PONG")
	return nil
}

// Foo the bar out of the blup.
func (*Bot) Foo(msg joe.Message) error {
	msg.Respond("FOO")
	return nil
}

func (*Bot) Hello(msg joe.Message) error {
	// Example of a function that does not have a documentation.
	msg.Respond("HELLO")
	return nil
}

// Global is an example of a global function.
// This is the second line.
//
// And another paragraph.
func Global(msg joe.Message) error {
	msg.Respond("GLOBAL")
	return nil
}
