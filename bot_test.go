package joe

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: test shutdown event context is not already canceled → Brain test
// TODO: test Bot.Respond
// TODO: test Bot.RespondRegex
// TODO: test Bot.Say

func TestBot_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	b := NewTest(ctx, t)

	initEvt := make(chan bool)
	b.Brain.RegisterHandler(func(evt InitEvent) {
		initEvt <- true
	})

	shutdownEvt := make(chan bool)
	b.Brain.RegisterHandler(func(evt ShutdownEvent) {
		shutdownEvt <- true
	})

	runExit := make(chan bool)
	go func() {
		assert.NoError(t, b.Run())
		runExit <- true
	}()

	wait(t, initEvt)
	cancel()

	wait(t, shutdownEvt)
	wait(t, runExit)
}

func TestBot_CloseAdapter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	input := &testCloser{Reader: new(bytes.Buffer)}
	output := new(bytes.Buffer)
	testAdapter := func(conf *Config) error {
		a := NewCLIAdapter("test", conf.Logger("adapter"))
		a.Input = input
		a.Output = output
		conf.SetAdapter(a)
		return nil
	}

	b := NewTest(ctx, t, testAdapter)

	runExit := make(chan bool)
	go func() {
		assert.NoError(t, b.Run())
		runExit <- true
	}()

	cancel()
	wait(t, runExit)
	assert.True(t, input.Closed)
}

func TestBot_ModuleErrors(t *testing.T) {
	ctx := context.Background()

	modA := func(conf *Config) error {
		return errors.New("error in module A")
	}

	modB := func(conf *Config) error {
		return errors.New("error in module B")
	}

	b := NewTest(ctx, t, modA, modB)

	err := b.Run()
	assert.EqualError(t, err, "failed to initialize bot: error in module A; error in module B")
}

func TestBot_RegistrationErrors(t *testing.T) {
	ctx := context.Background()
	b := NewTest(ctx, t)

	b.Brain.RegisterHandler(42)        // not a valid handler
	b.Brain.RegisterHandler(func() {}) // not a valid handler

	err := b.Run()
	require.Error(t, err)
	t.Log(err.Error())
	assert.Regexp(t, "invalid event handlers: .+", err.Error())
	assert.Regexp(t, "event handler is no function", err.Error())
	assert.Regexp(t, "event handler needs one or two arguments", err.Error())
}

type testCloser struct {
	Closed bool
	io.Reader
}

func (c *testCloser) Close() error {
	c.Closed = true
	return nil
}

func wait(t *testing.T, c chan bool) {
	select {
	case <-c:
		return
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
