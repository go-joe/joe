+++
title = "Minimal Example"
slug = "minimal"
weight = 1
pre = "<b>a) </b>"
+++

The simplest chat bot listens for messages on a chat _Adapter_ and then executes
a _Handler_ function if it sees a message directed to the bot that matches a given pattern.

For example a bot that responds to a message "ping" with the answer "PONG" looks like this:

[embedmd]:# (../../../_examples/01_minimal/main.go)
```go
package main

import "github.com/go-joe/joe"

func main() {
	b := joe.New("example-bot")
	b.Respond("ping", Pong)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func Pong(msg joe.Message) error {
	msg.Respond("PONG")
	return nil
}
```
