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
	T       TestingT
	Input   io.Writer
	Output  io.Reader
	Timeout time.Duration // defaults to 1s

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
		T:       t,
		Input:   input,
		Output:  output,
		Timeout: time.Second,
		runErr:  make(chan error, 1), // buffered so we can return from Bot.Run without blocking
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
	case <-time.After(b.Timeout):
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

	type InitTestEvent struct{}
	b.Brain.RegisterHandler(func(evt InitTestEvent) {
		started <- true
	})

	// When this event is handled we know the bot has completed its startup and
	// is ready to process events. The joe.InitEvent isn't really an option here
	// because it only marks that the bot is starting but we do not know when
	// all other init handlers are done (e.g. for the CLI adapter).
	b.Brain.Emit(InitTestEvent{})

	go func() {
		// The error will be available by calling Bot.Stop()
		err := b.Run()
		if err != nil {
			close(started)
		}
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go b.Brain.Shutdown(ctx)

	select {
	case err := <-b.runErr:
		if err != nil {
			b.T.Errorf("Bot.Run() returned an error: %v", err)
		}
	case <-time.After(b.Timeout):
		b.T.Errorf("Stop timed out")
		b.T.FailNow()
	}
}

// ReadOutput consumes all data from b.Output and returns it as a string so you
// can easily make assertions on it.
func (b *Bot) ReadOutput() string {
	out, err := ioutil.ReadAll(b.Output)
	if err != nil {
		b.T.Errorf("failed to read all output of bot: %v", err)
		return ""
	}

	return string(out)
}
