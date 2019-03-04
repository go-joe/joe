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

// TODO: test shutdown event context is not already canceled → Brain test
// TODO: test NewBrain uses in memory brain by default
// TODO: test Brain.Emit is asynchronous
// TODO: test HandleEvents
//       → InitEvent
//       → multiple handlers can match
//       → no handlers can match (e.g. wrong EventType)
//       → first external caller in registration errors
//       → passed context
//       → callbacks
//       → timeouts
//       → context done and shutdown event
// TODO: BrainMemoryEvents

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
			ctx, cancel := context.WithCancel(context.Background())
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
			go b.HandleEvents(ctx)
			defer cancel()

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

	ctx, cancel := context.WithCancel(context.Background())
	go b.HandleEvents(ctx)
	defer cancel()

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

	ctx, cancel := context.WithCancel(context.Background())
	go b.HandleEvents(ctx)
	defer cancel()

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

	ctx, cancel := context.WithCancel(context.Background())
	go b.HandleEvents(ctx)
	defer cancel()

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

	ctx, cancel := context.WithCancel(context.Background())
	go b.HandleEvents(ctx)
	defer cancel()

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

// EmitSync emits the given event on the brain and blocks until it has received
// the context which indicates that the event was fully processed by all
// matching handlers.
func EmitSync(b *Brain, event interface{}) {
	done := make(chan bool)
	callback := func(Event) { done <- true }
	b.Emit(event, callback)
	<-done
}
