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

// An Adapter connects the bot with the chat by enabling it to receive and send
// messages. Additionally advanced adapters can emit more events than just the
// ReceiveMessageEvent (e.g. the slack adapter also emits the UserTypingEvent).
// All adapter events must be setup in the Register function of the Adapter.
//
// Joe provides a default CLIAdapter implementation which connects the bot with
// the local shell to receive messages from stdin and print messages to stdout.
type Adapter interface {
	RegisterAt(*Brain)
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

// RegisterAt starts the CLIAdapter by reading messages from stdin and emitting
// a ReceiveMessageEvent for each of them. Additionally the adapter hooks into
// the InitEvent to print a nice prefix to stdout to show to the user it is
// ready to accept input.
func (a *CLIAdapter) RegisterAt(brain *Brain) {
	brain.RegisterHandler(func(evt InitEvent) {
		_ = a.print(a.Prefix)
	})

	go a.loop(brain)
}

func (a *CLIAdapter) loop(brain *Brain) {
	input := a.readLines()

	// The adapter loop is built to stay responsive even if the Brain stops
	// processing events so we can safely close the CLIAdapter.
	//
	// We want to print the prefix each time when the Brain has completely
	// processed a ReceiveMessageEvent and before we are emitting the next one.
	// This gives us a shell-like behavior which signals to the user that she
	// can input more data on the CLI. This channel is buffered so we do not
	// block the Brain when it executes the callback.
	callback := make(chan Event, 1)
	callbackFun := func(evt Event) {
		callback <- evt
	}

	var lines = input // channel represents the case that we receive a new message

	for {
		select {
		case msg, ok := <-lines:
			if !ok {
				// no more input from stdin
				lines = nil // disable this case and wait for closing signal
				continue
			}

			lines = nil // disable this case and wait for the callback
			brain.Emit(ReceiveMessageEvent{Text: msg}, callbackFun)

		case <-callback:
			// This case is executed after all ReceiveMessageEvent handlers have
			// completed and we can continue with the next line.
			_ = a.print(a.Prefix)
			lines = input // activate first case again

		case result := <-a.closing:
			if lines == nil {
				// We were just waiting for our callback
				_ = a.print(a.Prefix)
			}

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

	a.Logger.Debug("Closing CLIAdapter")
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
