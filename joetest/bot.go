package joetest

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"time"

	"github.com/go-joe/joe"
	"go.uber.org/zap/zaptest"
)

// Bot wraps a *joe.Bot for unit testing.
type Bot struct {
	*joe.Bot
	T      TestingT
	Input  io.Writer
	Output io.Reader

	runErr chan error
}

// NewBot creates a new *Bot instance that can be used in unit tests.
// The Bots will use a CLIAdapter which accepts messages from the Bot.Input
// and write all output to Bot.Output. The logger is a zaptest.Logger which
// sends all logs through the passed TestingT (usually a *testing.T instance).
//
// For ease of testing a Bot can be started and stopped without a cancel via
// Bot.Start() and Bot.Stop().
func NewBot(t TestingT, modules ...joe.Module) *Bot {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	input := new(bytes.Buffer)
	output := new(bytes.Buffer)

	b := &Bot{
		T:      t,
		Input:  input,
		Output: output,
		runErr: make(chan error, 1), // buffered so we can return from Bot.Run without blocking
	}

	testAdapter := joe.ModuleFunc(func(conf *joe.Config) error {
		a := joe.NewCLIAdapter("test", conf.Logger("adapter"))
		a.Input = ioutil.NopCloser(input)
		a.Output = output
		conf.SetAdapter(a)
		return nil
	})

	// The testAdapter and logger modules must be passed first so the caller can
	// actually inject a different Adapter or logger if required.
	testModules := []joe.Module{
		joe.WithLogger(logger),
		joe.WithContext(ctx),
		testAdapter,
	}

	b.Bot = joe.New("test", append(testModules, modules...)...)
	return b
}

// EmitSync emits the given event on the Brain and blocks until all registered
// handlers have completely processed it.
func (b *Bot) EmitSync(event interface{}) {
	b.T.Helper()

	done := make(chan bool)
	callback := func(joe.Event) { done <- true }
	b.Brain.Emit(event, callback)

	select {
	case <-done:
		// ok, cool
	case <-time.After(time.Second):
		b.T.Errorf("EmitSync timed out")
		b.T.FailNow()
	}
}

// Start executes the Bot.Run() function and stores its error result in a channel
// so the caller can eventually execute Bot.Stop() and receive the result.
// This function blocks until the event handler is actually running and emits
// the InitEvent.
func (b *Bot) Start() {
	started := make(chan bool)
	b.Brain.RegisterHandler(func(evt joe.InitEvent) {
		started <- true
	})

	go func() {
		// The error will be available by calling Bot.Stop()
		_ = b.Run()
	}()

	<-started
}

// Run wraps Bot.Run() in order to allow stopping a Bot without having to
// inject another context.
func (b *Bot) Run() error {
	b.T.Helper()
	err := b.Bot.Run()
	b.runErr <- err // b.runErr is buffered so we can return immediately
	return err
}

// Stop stops a running Bot and blocks until it has completed. If Bot.Run()
// returned an error it is passed to the Errorf function of the TestingT that
// was used to create the Bot.
func (b *Bot) Stop() {
	ctx := context.Background()
	b.Brain.Shutdown(ctx)
	err := <-b.runErr
	if err != nil {
		b.T.Errorf("Bot.Run() returned an error: %v", err)
	}
}
