package joe

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"

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
	Input  io.Writer
	Output io.Reader
}

// NewTest creates a new *Bot instance that can be used in unit tests.
// The Bot will
func NewTest(ctx context.Context, t TestingT, modules ...Module) *TestBot {
	logger := zaptest.NewLogger(t)
	input := new(bytes.Buffer)
	output := new(bytes.Buffer)

	b := &TestBot{
		Input:  input,
		Output: output,
	}

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
