package joe

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// An adapter connects the bot with the chat by enabling it to receive and send
// messages. Additionally advanced adapters can emit more events than just the
// ReceiveMessageEvent (e.g. the slack adapter also emits the UserTypingEvent).
// All adapter events must be setup in the Register function of the Adapter.
//
// Joe provides a default CLIAdapter implementation which connects the bot with
// the local shell to receive messages from stdin and print messages to stdout.
type Adapter interface {
	Register(Events)
	Send(text, channel string) error
	Close() error
}

// The CLIAdapter is the default Adapter implementation that the bot uses if no
// other adapter was configured. It emits a ReceiveMessageEvent for each line it
// receives from stdin and prints all sent messages to stdout.
type CLIAdapter struct {
	Prefix  string
	Input   io.ReadCloser
	Output  io.Writer
	Logger  *zap.Logger
	mu      sync.Mutex // protects the Output and closing channel
	closing chan chan error
}

// NewCLIAdapter creates a new CLIAdapter. The caller must call Close
// to make the CLIAdapter stop reading messages and emitting events.
func NewCLIAdapter(name string, logger *zap.Logger) *CLIAdapter {
	return &CLIAdapter{
		Prefix:  fmt.Sprintf("%s > ", name),
		Input:   os.Stdin,
		Output:  os.Stdout,
		Logger:  logger,
		closing: make(chan chan error),
	}
}

// Register starts the CLIAdapter by reading messages from stdin and emitting a
// ReceiveMessageEvent for each of them. Additionally the adapter hooks into the
// InitEvent to print a nice prefix to stdout to show to the user it is ready to
// accept input.
func (a *CLIAdapter) Register(events Events) {
	events.RegisterHandler(func(evt InitEvent) {
		_ = a.print(a.Prefix)
	})

	go a.loop(events.Channel())
}

func (a *CLIAdapter) loop(events chan<- event) {
	callback := func(event) {
		// We want to print the prefix each time we are done with handling a message.
		_ = a.print(a.Prefix)
	}

	input := a.readLines()

	var (
		lines = input      // channel represents the case that we receive a new message
		emit  chan<- event // channel to activate the case that the event was delivered
		evt   event        // the event to deliver (if any)
	)

	for {
		select {
		case msg, ok := <-lines:
			if !ok {
				// no more input from stdin
				lines = nil // disable this case and wait for closing signal
				continue
			}

			lines = nil   // disable this case
			emit = events // enable the event delivery case
			evt = event{data: ReceiveMessageEvent{Text: msg}}
			evt.callbacks = append(evt.callbacks, callback)

		case emit <- evt:
			emit = nil    // disable this case
			lines = input // activate first case again
			evt = event{} // release old event data

		case result := <-a.closing:
			_ = a.print("\n")
			result <- a.Input.Close()
			return
		}
	}
}

// ReadLines reads lines from stdin and returns them in a channel.
// All strings in the returned channel will not include the trailing newline.
// The channel is closed automatically when a.Input is closed.
func (a *CLIAdapter) readLines() <-chan string {
	r := bufio.NewReader(a.Input)
	lines := make(chan string)
	go func() {
		// This goroutine will exit when we call a.Input.Close() which will make
		// r.ReadString(…) return an io.EOF.
		for {
			line, err := r.ReadString('\n')
			switch {
			case err == io.EOF:
				close(lines)
				return
			case err != nil:
				a.Logger.Error("Failed to read messages from input", zap.Error(err))
				return
			}

			lines <- line[:len(line)-1]
		}
	}()

	return lines
}

// Send implements the Adapter interface by sending the given text to stdout.
// The channel argument is required by the Adapter interface but is otherwise ignored.
func (a *CLIAdapter) Send(text, channel string) error {
	return a.print(text + "\n")
}

// Close makes the CLIAdapter stop emitting any new events or printing any output.
// Calling this function more than once will result in an error.
func (a *CLIAdapter) Close() error {
	if a.closing == nil {
		return errors.Errorf("already closed")
	}

	callback := make(chan error)
	a.closing <- callback
	err := <-callback

	// Mark CLIAdapter as closed by setting its closing channel to nil.
	// This will prevent any more output to be printed after this function returns.
	a.mu.Lock()
	a.closing = nil
	a.mu.Unlock()

	return err
}

func (a *CLIAdapter) print(msg string) error {
	a.mu.Lock()
	if a.closing == nil {
		return errors.New("adapter is closed")
	}
	_, err := fmt.Fprint(a.Output, msg)
	a.mu.Unlock()

	return err
}
