package joe

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
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
	Input  io.Reader
	Output io.Writer
	Logger *zap.Logger
	ctx    context.Context
}

// NewCLIAdapter creates a new CLIAdapter. The caller must call Close
// to make the CLIAdapter stop reading messages and emitting events.
func NewCLIAdapter(ctx context.Context, name string, logger *zap.Logger) *CLIAdapter {
	return &CLIAdapter{
		Prefix: fmt.Sprintf("%s > ", name),
		Input:  os.Stdin,
		Output: os.Stdout,
		Logger: logger,
		ctx:    ctx,
	}
}

// Register starts the CLIAdapter by reading messages from stdin and emitting a
// ReceiveMessageEvent for each of them. Additionally the adapter hooks into the
// InitEvent to print a nice prefix to stdout to show to the user it is ready to
// accept input.
func (a *CLIAdapter) Register(b *Brain) {
	b.RegisterHandler(func(evt InitEvent) {
		a.print(a.Prefix)
	})

	go func() {
		callback := func(event) {
			a.print(a.Prefix)
		}

		for line := range a.readLines() {
			b.emit(ReceiveMessageEvent{Text: line}, callback)
		}
	}()
}

// ReadLines reads lines from stdin and returns them in a channel.
// All strings in the returned channel will not include the trailing newline.
// The channel is closed automatically if there are no more lines or if the
// context is closed.
func (a *CLIAdapter) readLines() <-chan string {
	r := bufio.NewReader(a.Input)
	c := make(chan string)
	go func() {
		// TODO: make sure this isnt leaked?
		for {
			line, err := r.ReadString('\n')
			switch {
			case err == io.EOF:
				close(c)
				return
			case err != nil:
				a.Logger.Error("Failed to read messages from input", zap.Error(err))
				return
			}

			c <- line[:len(line)-1]
		}
	}()

	lines := make(chan string)
	go func() {
		for {
			select {
			case l, ok := <-c:
				if !ok {
					close(lines)
					return
				}
				lines <- l
			case <-a.ctx.Done():
				close(lines)
				return
			}
		}
	}()

	return lines
}

// Send implements the Adapter interface by sending the given text to stdout.
// The channel argument is required by the Adapter interface but is otherwise ignored.
func (a *CLIAdapter) Send(text, channel string) error {
	a.println(text)
	return nil
}

// Close implements the Adapter interface simply by printing a final newline to stdout.
func (a *CLIAdapter) Close() error {
	a.println()
	return nil
}

func (a *CLIAdapter) print(msg string) {
	_, _ = fmt.Fprint(a.Output, msg)
}

func (a *CLIAdapter) println(msg ...interface{}) {
	_, _ = fmt.Fprintln(a.Output, msg...)
}
