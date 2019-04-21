package main

import (
	"context"
	"errors"

	joehttp "github.com/go-joe/http-server"
	"github.com/go-joe/joe"
)

type ExampleBot struct {
	*joe.Bot
}

func main() {
	b := &ExampleBot{Bot: joe.New("example",
		joehttp.Server(":8080"),
	)}

	b.Brain.RegisterHandler(b.HandleHTTP)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func (b *ExampleBot) HandleHTTP(context.Context, joehttp.RequestEvent) error {
	return errors.New("TODO: Add your custom logic here")
}
