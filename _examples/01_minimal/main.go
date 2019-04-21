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
