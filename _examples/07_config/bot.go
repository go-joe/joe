package main

import (
	joehttp "github.com/go-joe/http-server"
	"github.com/go-joe/joe"
	"github.com/pkg/errors"
)

type Bot struct {
	*joe.Bot        // Anonymously embed the joe.Bot type so we can use its functions easily.
	conf     Config // You can keep other fields here as well.
}

func New(conf Config) *Bot {
	b := &Bot{
		Bot: joe.New("joe", conf.Modules()...),
	}

	// Define any custom event and message handlers here
	b.Brain.RegisterHandler(b.GitHubCallback)
	b.Respond("do stuff", b.DoStuffCommand)

	return b
}

func New2(conf Config) (*Bot, error) {
	if err := conf.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	b := &Bot{
		Bot: joe.New("joe", conf.Modules()...),
	}

	// Define any custom event and message handlers here
	b.Brain.RegisterHandler(b.GitHubCallback)
	b.Respond("do stuff", b.DoStuffCommand)

	return b, nil
}

func (b *Bot) GitHubCallback(joehttp.RequestEvent) {
	// Handler only provided for completeness.
}

func (b *Bot) DoStuffCommand(joe.Message) error {
	// Handler only provided for completeness.
	return nil
}
