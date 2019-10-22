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
