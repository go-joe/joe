package joe

import (
	"context"
	"fmt"

	"github.com/fraugster/cli"
)

type CLIAdapter struct {
	Prefix   string
	Messages <-chan string
	ctx      context.Context
}

func NewCLIAdapter(ctx context.Context, name string) *CLIAdapter {
	return &CLIAdapter{
		Prefix:   fmt.Sprintf("%s > ", name),
		Messages: cli.ReadLines(ctx),
		ctx:      ctx,
	}
}

func (a *CLIAdapter) NextMessage() <-chan string {
	select {
	case <-a.ctx.Done():
		return a.Messages
	default:
		fmt.Print(a.Prefix)
		return a.Messages
	}
}

func (*CLIAdapter) Send(msg string) error {
	fmt.Println(msg)
	return nil
}

func (*CLIAdapter) Close() error {
	fmt.Println()
	return nil
}
