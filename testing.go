package joe

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"time"

	"go.uber.org/zap/zaptest"
)

// TestingT is the minimal required subset of the API provided by all *testing.T
// and *testing.B objects.
type TestingT interface {
	Logf(string, ...interface{})
	Errorf(string, ...interface{})
	Fail()
	Failed() bool
	Name() string
	FailNow()
}

// TestBot wraps a *Bot for unit tests.
type TestBot struct {
	*Bot
	T      TestingT
	Input  io.Writer
	Output io.Reader

	stop   func()
	runErr chan error
}

// NewTest creates a new *Bot instance that can be used in unit tests.
// The Bots will use a CLIAdapter which accepts messages from the TestBot.Input
// and write all output to TestBot.Output. The logger is a zaptest.Logger which
// sends all logs through the passed TestingT (usually a *testing.T instance).
//
// For ease of testing a TestBot can be started and stopped without a cancel via
// TestBot.Start() and TestBot.Stop().
func NewTest(t TestingT, modules ...Module) *TestBot {
	logger := zaptest.NewLogger(t)
	input := new(bytes.Buffer)
	output := new(bytes.Buffer)

	b := &TestBot{
		T:      t,
		Input:  input,
		Output: output,
		runErr: make(chan error, 1), // buffered so we can return from Bot.Run without blocking
	}

	ctx := context.Background()
	ctx, b.stop = context.WithCancel(ctx)

	testAdapter := func(conf *Config) error {
		a := NewCLIAdapter("test", conf.Logger("adapter"))
		a.Input = ioutil.NopCloser(input)
		a.Output = output
		conf.SetAdapter(a)
		return nil
	}

	// The testAdapter module must be first so the caller can actually inject a
	// different Adapter if required.
	modules = append([]Module{testAdapter}, modules...)
	b.Bot = newBot(ctx, logger, "test", modules...)
	return b
}

// EmitSync emits the given event on the Brain and blocks until all registered
// handlers have completely processed it.
func (b *TestBot) EmitSync(t TestingT, event interface{}) {
	done := make(chan bool)
	callback := func(Event) { done <- true }
	b.Brain.Emit(event, callback)

	select {
	case <-done:
		// ok, cool
	case <-time.After(time.Second):
		t.Errorf("timeout")
	}
}

// Start executes the Bot.Run() function and stores its error result in a channel
// so the caller can eventually execute TestBot.Stop() and receive the result.
func (b *TestBot) Start() {
	go func() {
		// The error will be available by calling TestBot.Stop()
		_ = b.Run()
	}()
}

// Run wraps Bot.Run() in order to allow stopping a TestBot without having to
// inject another context.
func (b *TestBot) Run() error {
	err := b.Bot.Run()
	b.runErr <- err // b.runErr is buffered so we can return immediately
	return err
}

// Stop stops a running TestBot and blocks until it has completed. If Bot.Run()
// returned an error it is passed to the Errorf function of the TestingT that
// was used to create the TestBot.
func (b *TestBot) Stop() {
	b.stop()
	err := <-b.runErr
	if err != nil {
		b.T.Errorf("Bot.Run() returned an error: %v")
	}
}
