package joe

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TODO: test shutdown event context is not already canceled → Brain test
// TODO: test NewBrain uses in memory brain by default
// TODO: test Brain.Emit is asynchronous
// TODO: test HandleEvents
//       → InitEvent
//       → multiple handlers can match
//       → no handlers can match (e.g. wrong EventType)
//       → returned errors
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
