package joe

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func TestBrain_RegisterHandler(t *testing.T) {
	type TestEvent struct {
		EventHandled *sync.WaitGroup
	}

	cases := map[string]struct {
		fun interface{}
		err string
	}{
		"err_no_arg": {
			fun: func() {},
			err: "event handler needs one or two arguments",
		},
		"err_pointer": {
			fun: func(evt *TestEvent) {},
			err: "event handler argument must be a struct and not a pointer",
		},
		"err_too_many_args": {
			fun: func(evt1, evt2, evt3, evt4 TestEvent) {},
			err: "event handler needs one or two arguments",
		},
		"err_too_many_events": {
			fun: func(evt1, evt2 TestEvent) {},
			err: "event handler has two arguments but the first is not a context.Context",
		},
		"err_wrong_arg": {
			fun: func(n int) {},
			err: "event handler argument must be a struct",
		},
		"err_context": {
			fun: func(TestEvent, context.Context) {},
			err: "event handler context must be the first argument",
		},
		"err_too_many_results": {
			fun: func(TestEvent) (err1, err2 error) { return nil, nil },
			err: "event handler has more than one return value",
		},
		"err_wrong_result": {
			fun: func(TestEvent) int { return 42 },
			err: "if the event handler has a return value it must implement the error interface",
		},
		"ok_simple": {
			fun: func(evt TestEvent) {
				evt.EventHandled.Done()
			},
		},
		"ok_with_error": {
			fun: func(evt TestEvent) error {
				evt.EventHandled.Done()
				return nil
			},
		},
		"ok_with_context": {
			fun: func(ctx context.Context, evt TestEvent) {
				evt.EventHandled.Done()
			},
		},
		"ok_with_context_and_error": {
			fun: func(ctx context.Context, evt TestEvent) error {
				evt.EventHandled.Done()
				return nil
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			b := NewBrain(logger)
			b.RegisterHandler(c.fun)

			if c.err != "" {
				require.Len(t, b.registrationErrs, 1)
				err := b.registrationErrs[0].Error()
				if !strings.HasSuffix(err, c.err) {
					t.Errorf("unexpected registration error\nexpected: %q\nactual  : %q", c.err, err)
				}
				return
			}

			require.Empty(t, b.registrationErrs, "unexpected registration errors")

			// Start the brains event handler loop.
			go b.HandleEvents()
			defer b.Shutdown()

			// Emit our test event.
			wg := new(sync.WaitGroup)
			wg.Add(1)
			evt := TestEvent{EventHandled: wg}
			b.Emit(evt)

			// Wait until the handler marks the event as handled
			done := make(chan bool)
			go func() {
				wg.Wait()
				done <- true
			}()

			select {
			case <-done:
				// ok cool
			case <-time.After(time.Second):
				t.Error("Event handler was not executed within one second")
			}
		})
	}
}

func TestBrain_HandlerErrors(t *testing.T) {
	type TestEvent struct{}

	// In this test we actually want to check if the handler errors get logged.
	// This is achieved by using go.uber.org/zap/zaptest/observer
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	b := NewBrain(logger)

	handlerErr := errors.New("test error")
	b.RegisterHandler(func(TestEvent) error {
		return handlerErr
	})

	go b.HandleEvents()
	defer b.Shutdown()

	EmitSync(b, TestEvent{})

	expectedLog := observer.LoggedEntry{
		Entry:   zapcore.Entry{Level: zap.ErrorLevel, Message: "Event handler failed"},
		Context: []zapcore.Field{zap.Error(handlerErr)},
	}

	handlerErrLogs := logs.FilterMessage(expectedLog.Message).AllUntimed()
	require.Equal(t, 1, len(handlerErrLogs))
	assert.Equal(t, expectedLog, handlerErrLogs[0])
}

func TestBrain_Emit_PassAllEventData(t *testing.T) {
	type TestEvent struct {
		Test       bool
		unexported string
	}

	logger := zaptest.NewLogger(t)
	b := NewBrain(logger)

	var seen TestEvent
	b.RegisterHandler(func(evt TestEvent) {
		seen = evt
	})

	go b.HandleEvents()
	defer b.Shutdown()

	event := TestEvent{Test: true, unexported: "hello"}
	EmitSync(b, event)

	assert.Equal(t, event, seen)
}

func TestBrain_Emit_ImmutableEvent(t *testing.T) {
	type TestEvent struct {
		String string
	}

	logger := zaptest.NewLogger(t)
	b := NewBrain(logger)

	b.RegisterHandler(func(evt TestEvent) {
		evt.String = "bar"
	})

	go b.HandleEvents()
	defer b.Shutdown()

	event := TestEvent{String: "foo"}
	EmitSync(b, event)

	assert.Equal(t, "foo", event.String)
}

func TestBrain_HandlerPanics(t *testing.T) {
	type TestEvent struct{}

	// In this test we actually want to check if the handler panic gets logged.
	// This is achieved by using go.uber.org/zap/zaptest/observer
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	b := NewBrain(logger)

	var handlerCalled bool
	b.RegisterHandler(func(TestEvent) {
		handlerCalled = true
		panic("something went horribly wrong")
	})

	go b.HandleEvents()
	defer b.Shutdown()

	EmitSync(b, TestEvent{})
	assert.True(t, handlerCalled)

	handlerErrLogs := logs.FilterMessage("Event handler failed")
	require.Equal(t, 1, handlerErrLogs.Len())
	logEntry := handlerErrLogs.All()[0]
	assert.Equal(t, "error", logEntry.Level.String())
	assert.NotEmpty(t, logEntry.Context, "expected log entry to have at least one field")
	for _, field := range logEntry.Context {
		switch field.Key {
		case "error":
			assert.Equal(t, zapcore.ErrorType, field.Type)
			err := field.Interface.(error)
			assert.EqualError(t, err, "handler panic: something went horribly wrong")
		default:
			t.Errorf("unexpected field %q in log entry", field.Key)
		}
	}
}

func TestBrain_Memory(t *testing.T) {
	logger := zaptest.NewLogger(t)
	b := NewBrain(logger)

	var events []BrainMemoryEvent
	b.RegisterHandler(func(evt BrainMemoryEvent) {
		events = append(events, evt)
	})

	go b.HandleEvents()

	require.NoError(t, b.Set("foo", "bar"))
	require.NoError(t, b.Set("hello", "world"))

	val, ok, err := b.Get("foo")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "bar", val)

	mem, err := b.Memories()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"foo": "bar", "hello": "world"}, mem)

	ok, err = b.Delete("hello")
	require.NoError(t, err)
	assert.True(t, ok)

	b.Shutdown()

	expectedEvents := []BrainMemoryEvent{
		{Operation: "set", Key: "foo", Value: "bar"},
		{Operation: "set", Key: "hello", Value: "world"},
		{Operation: "get", Key: "foo", Value: "bar"},
		{Operation: "del", Key: "hello"},
	}

	assert.Equal(t, expectedEvents, events)
}

func TestBrain_Shutdown_WithoutStart(t *testing.T) {
	logger := zaptest.NewLogger(t)
	b := NewBrain(logger)

	done := make(chan bool)
	go func() {
		b.Shutdown()
		done <- true
	}()

	select {
	case <-done:
		// hurray!
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestBrain_Shutdown_MultipleTimes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	b := NewBrain(logger)

	n := 100
	done := make(chan bool, n)
	for i := 0; i < n; i++ {
		go func() {
			b.Shutdown()
			done <- true
		}()
	}

	// All shutdown functions should return and nothing should deadlock or cause
	// a panic (e.g. closing channels twice).
	for i := 0; i < n; i++ {
		select {
		case <-done:
			// hurray!
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}
}

func TestBrain_EmitAfterShutdown(t *testing.T) {
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)
	b := NewBrain(logger)

	b.Shutdown()

	// Emitting new events after shutdown should not block or panic
	type TestEvent struct{}

	b.Emit(ReceiveMessageEvent{})
	b.Emit(UserTypingEvent{})
	b.Emit(TestEvent{})

	all := logs.AllUntimed()
	require.Len(t, all, 3)
	for i, logEvent := range all {
		assert.Equal(t, "Ignoring new event because brain is currently shutting down or is already closed", logEvent.Message)
		require.Len(t, logEvent.Context, 1)
		assert.Equal(t, "type", logEvent.Context[0].Key)
		switch i {
		case 0:
			assert.Equal(t, "joe.ReceiveMessageEvent", logEvent.Context[0].String)
		case 1:
			assert.Equal(t, "joe.UserTypingEvent", logEvent.Context[0].String)
		case 2:
			assert.Equal(t, "joe.TestEvent", logEvent.Context[0].String)
		}
	}
}

// EmitSync emits the given event on the brain and blocks until it has received
// the context which indicates that the event was fully processed by all
// matching handlers.
func EmitSync(brain EventEmitter, event interface{}) {
	done := make(chan bool)
	callback := func(Event) { done <- true }
	brain.Emit(event, callback)
	<-done
}
