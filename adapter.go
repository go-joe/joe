package joe

import (
	"context"
	"fmt"

	"github.com/fraugster/cli"
)

// An adapter connects the bot with the chat by enabling it to receive and send
// messages. Additionally advanced adapters can emit more events than just the
// ReceiveMessageEvent (e.g. the slack adapter also emits the UserTypingEvent).
// Such adapter events must be setup in the Register function of the Adapter.
//
// Joe provides a default CLIAdapter implementation which connects the bot with
// the local shell to receive messages from stdin and print messages to stdout.
type Adapter interface {
	Register(*Brain)
	Send(text, channel string) error
	Close() error
}

// The CLIAdapter is the default Adapter implementation that the bot uses if no
// other adapter was configured. It emits a ReceiveMessageEvent for each line it
// receives from stdin and prints all sent messages to stdout.
type CLIAdapter struct {
	Prefix string
	ctx    context.Context
}

// NewCLIAdapter creates a new CLIAdapter.
// The passed context is used to close the channel that is receiving messages
// from stdin.
func NewCLIAdapter(ctx context.Context, name string) *CLIAdapter {
	return &CLIAdapter{
		Prefix: fmt.Sprintf("%s > ", name),
		ctx:    ctx,
	}
}

// Register starts the CLIAdapter by reading messages from stdin and emitting a
// ReceiveMessageEvent for each of them. Additionally the adapter hooks into the
// InitEvent to print a nice prefix to stdout to show to the user it is ready to
// accept input.
func (a *CLIAdapter) Register(b *Brain) {
	b.RegisterHandler(func(evt InitEvent) {
		fmt.Print(a.Prefix)
	})

	go func() {
		callback := func(event) {
			fmt.Print(a.Prefix)
		}

		for line := range cli.ReadLines(a.ctx) {
			b.Emit(ReceiveMessageEvent{Text: line}, callback)
		}
	}()
}

// Send implements the Adapter interface by sending the given text to stdout.
// The channel argument is required by the Adapter interface but is otherwise ignored.
func (*CLIAdapter) Send(text, channel string) error {
	fmt.Println(text)
	return nil
}

// Close implements the Adapter interface simply by printing a final newline to stdout.
func (*CLIAdapter) Close() error {
	fmt.Println()
	return nil
}
