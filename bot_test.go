package joe

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestBot_Run(t *testing.T) {
	b := NewTest(t)

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
	b.Stop()

	wait(t, shutdownEvt)
	wait(t, runExit)
}

func TestBot_Respond(t *testing.T) {
	b := NewTest(t)
	handledMessages := make(chan Message)
	b.Respond("Hello (.+), this is a (.+)", func(msg Message) error {
		handledMessages <- msg
		return nil
	})

	b.Start()
	defer b.Stop()

	b.Brain.Emit(ReceiveMessageEvent{
		Text:    "Hello world, this is a test",
		Channel: "XXX",
	})

	select {
	case msg := <-handledMessages:
		assert.Equal(t, "Hello world, this is a test", msg.Text)
		assert.Equal(t, "XXX", msg.Channel)
		assert.Equal(t, []string{"world", "test"}, msg.Matches)
	case <-time.After(time.Second):
		t.Error("Timeout")
	}
}

func TestBot_Respond_Matches(t *testing.T) {
	b := NewTest(t)
	handledMessages := make(chan Message)
	b.Respond("Remember (.+) is (.+)", func(msg Message) error {
		handledMessages <- msg
		return nil
	})

	b.Start()
	defer b.Stop()

	cases := map[string][]string{
		"Remember foo is bar": {"foo", "bar"},
		"remember a is b":     {"a", "b"},
		"remember FOO IS BAR": {"FOO", "BAR"},
	}

	for input, matches := range cases {
		b.Brain.Emit(ReceiveMessageEvent{Text: input})
		select {
		case msg := <-handledMessages:
			assert.Equal(t, matches, msg.Matches)
		case <-time.After(time.Second):
			t.Error("timeout")
		}
	}
}

func TestBot_Respond_No_Matches(t *testing.T) {
	b := NewTest(t)
	b.Respond("Hello world, this is a test", func(msg Message) error {
		t.Errorf("Handler should not match but got %+v", msg)
		return nil
	})

	nonMatches := []string{
		"Foobar",                                // entirely different
		"Hello world",                           // only the prefix
		"this is a test",                        // only the suffix
		"world",                                 // only a substring
		"Hello world this is a test",            // missing comma
		"TEST Hello world, this is a test",      // additional prefix
		"Hello world, this is a test TEST",      // additional suffix
		"TEST Hello world, this is a test TEST", // additional prefix and suffix
		"Hello world, TEST this is a test",      // additional word in the middle
	}

	b.Start()
	defer b.Stop()

	for _, txt := range nonMatches {
		b.EmitSync(t, ReceiveMessageEvent{Text: txt})
	}
}

func TestBot_RespondRegex(t *testing.T) {
	b := NewTest(t)
	handledMessages := make(chan Message, 1)
	b.RespondRegex(`name is ([^\s]+)$`, func(msg Message) error {
		t.Logf("Received message %q", msg.Text)
		handledMessages <- msg
		return nil
	})

	b.Start()
	defer b.Stop()

	cases := map[string][]string{ // maps input to expected matches
		"name is Joe":                       {"Joe"}, // simple case
		"NAME IS Joe":                       {"Joe"}, // simple case, case insensitive
		"Hello, my name is Joe":             {"Joe"}, // match on substrings
		"My name is Joe and what is yours?": nil,     // respect end of input anchor
		"":                                  nil,     // should not match but also not panic
	}

	for input, matches := range cases {
		b.EmitSync(t, ReceiveMessageEvent{Text: input})

		if matches == nil {
			select {
			case msg := <-handledMessages:
				t.Errorf("message handler should not have been called with %q", msg.Text)
				continue
			default:
				// no message as expected, lets move on
				continue
			}
		}

		// Check message was handled as expected
		select {
		case msg := <-handledMessages:
			assert.Equal(t, matches, msg.Matches)
		case <-time.After(time.Second):
			t.Errorf("timeout: %s", input)
		}
	}
}

func TestBot_RespondRegex_Empty(t *testing.T) {
	b := NewTest(t)
	b.RespondRegex("", func(msg Message) error {
		t.Error("should never match")
		return nil
	})

	b.Start()
	defer b.Stop()

	cases := []string{
		"",
		"   ",
		"\n",
		"\t",
		"foobar",
		"foo bar",
	}

	for _, input := range cases {
		b.EmitSync(t, ReceiveMessageEvent{Text: input})
	}
}

func TestBot_RespondRegex_Invalid(t *testing.T) {
	b := NewTest(t)
	b.RespondRegex("this is not a [valid regular expression", func(msg Message) error {
		t.Error("should never match")
		return nil
	})

	err := b.Run()
	require.EqualError(t, err, "invalid event handlers: failed to add Response handler: "+
		"error parsing regexp: missing closing ]: `[valid regular expression`")
}

func TestBot_CloseAdapter(t *testing.T) {
	input := &testCloser{Reader: new(bytes.Buffer)}
	output := new(bytes.Buffer)
	testAdapter := func(conf *Config) error {
		a := NewCLIAdapter("test", conf.Logger("adapter"))
		a.Input = input
		a.Output = output
		conf.SetAdapter(a)
		return nil
	}

	b := NewTest(t, testAdapter)

	b.Start()
	b.Stop()

	assert.True(t, input.Closed)
}

func TestBot_ModuleErrors(t *testing.T) {
	modA := func(conf *Config) error {
		return errors.New("error in module A")
	}

	modB := func(conf *Config) error {
		return errors.New("error in module B")
	}

	b := NewTest(t, modA, modB)

	err := b.Run()
	assert.EqualError(t, err, "failed to initialize bot: error in module A; error in module B")
}

func TestBot_RegistrationErrors(t *testing.T) {
	b := NewTest(t)

	b.Brain.RegisterHandler(42)        // not a valid handler
	b.Brain.RegisterHandler(func() {}) // not a valid handler

	err := b.Run()
	require.Error(t, err)
	t.Log(err.Error())
	assert.Regexp(t, "invalid event handlers: .+", err.Error())
	assert.Regexp(t, "event handler is no function", err.Error())
	assert.Regexp(t, "event handler needs one or two arguments", err.Error())
}

// TestBot_Logger simply tests that the zap logger configuration in newLogger()
// doesn't panic.
func TestBot_Logger(t *testing.T) {
	newLogger()
}

func TestBot_Say(t *testing.T) {
	a := new(MockAdapter)
	b := NewTest(t)
	b.Adapter = a

	a.On("Send", "Hello world", "foo").Return(nil)
	b.Say("foo", "Hello world")

	a.On("Send", "Hello world: the answer is 42", "bar").Return(nil)
	b.Say("bar", "Hello %s: the answer is %d", "world", 42)

	a.AssertExpectations(t)
}

func TestBot_Say_Error(t *testing.T) {
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	a := new(MockAdapter)
	b := NewTest(t)
	b.Adapter = a
	b.Logger = logger

	adapterErr := errors.New("watch your language")
	a.On("Send", "damn it", "baz").Return(adapterErr)
	b.Say("baz", "damn it")

	assert.Equal(t, []observer.LoggedEntry{{
		Entry:   zapcore.Entry{Level: zap.ErrorLevel, Message: "Failed to send message"},
		Context: []zapcore.Field{zap.Error(adapterErr)},
	}}, logs.AllUntimed())

	a.AssertExpectations(t)
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

type MockAdapter struct {
	mock.Mock
}

func (a *MockAdapter) Register(r EventRegistry) {
	a.Called(r)
}

func (a *MockAdapter) Send(text, channel string) error {
	args := a.Called(text, channel)
	return args.Error(0)
}

func (a *MockAdapter) Close() error {
	args := a.Called()
	return args.Error(0)
}
