package joe

import (
	"context"
	"fmt"

	"github.com/fraugster/cli"
)

type Adapter interface {
	Register(*EventProcessor)
	Send(text, channelID string) error
	Close() error
}

type CLIAdapter struct {
	Prefix string
	ctx    context.Context
}

func NewCLIAdapter(ctx context.Context, name string) *CLIAdapter {
	return &CLIAdapter{
		Prefix: fmt.Sprintf("%s > ", name),
		ctx:    ctx,
	}
}

func (a *CLIAdapter) Register(p *EventProcessor) {
	p.RegisterHandler(func(evt InitEvent) {
		fmt.Print(a.Prefix)
	})

	go func() {
		callback := func(Event) {
			fmt.Print(a.Prefix)
		}

		for line := range cli.ReadLines(a.ctx) {
			p.Emit(ReceiveMessageEvent{Text: line}, callback)
		}
	}()
}

func (*CLIAdapter) Send(text, _ string) error {
	fmt.Println(text)
	return nil
}

func (*CLIAdapter) Close() error {
	fmt.Println()
	return nil
}
