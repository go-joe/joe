package joetest

import (
	"context"
	"sync"
	"time"

	"github.com/go-joe/joe"
	"go.uber.org/zap/zaptest"
)

// Brain wraps the joe.Brain for unit testing.
type Brain struct {
	*joe.Brain

	mu     sync.Mutex
	events []interface{}
}

// NewBrain creates a new Brain that can be used for unit testing. The Brain
// registers to all events except the (init and shutdown event) and records them
// for later access. The event handling loop of the Brain (i.e. Brain.HandleEvents())
// is automatically started by this function in a new goroutine and the caller
// must call Brain.Finish() at the end of their tests.
func NewBrain(t TestingT) *Brain {
	logger := zaptest.NewLogger(t)
	b := &Brain{Brain: joe.NewBrain(logger)}

	initialized := make(chan bool)
	b.RegisterHandler(b.observeEvent)
	b.RegisterHandler(func(joe.InitEvent) {
		initialized <- true
	})

	go b.HandleEvents()
	<-initialized

	return b
}

func (b *Brain) observeEvent(evt interface{}) {
	switch evt.(type) {
	case joe.InitEvent, joe.ShutdownEvent:
		return
	default:
		b.mu.Lock()
		b.events = append(b.events, evt)
		b.mu.Unlock()
	}
}

// Finish stops the event handler loop of the Brain and waits until all pending
// events have been processed.
func (b *Brain) Finish() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	b.Brain.Shutdown(ctx)
}

// RecordedEvents returns all events the Brain has processed except the
// joe.InitEvent and joe.ShutdownEvent.
func (b *Brain) RecordedEvents() []interface{} {
	b.mu.Lock()
	events := make([]interface{}, len(b.events))
	copy(events, b.events)
	b.mu.Unlock()

	return events
}
