package joe

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// The Brain contains the core logic of a Bot by implementing an event handler
// that dispatches events to all registered event handlers. Additionally the
// Brain is directly connected to the Memory of the bot to manage concurrent
// access as well as to emit the BrainMemoryEvent if memory is created, edited
// or deleted on the brain.
type Brain struct {
	logger *zap.Logger

	mu     sync.RWMutex // mu protects concurrent access to the Memory
	memory Memory

	events         chan Event
	handlers       map[reflect.Type][]eventHandler
	handlerTimeout time.Duration // zero means no timeout

	registrationErrs []error // any errors that occurred during setup (e.g. in Bot.RegisterHandler)
}

// An Event represents a concrete event type and optional callbacks that are
// triggered when the event was processed by any handler.
type Event struct {
	Data      interface{}
	Callbacks []func(Event)
}

// An event handler is a function that takes a context and the reflected value
// of a concrete event type.
type eventHandler func(context.Context, reflect.Value) error

// The EventRegistry is the interface that is exposed to Adapter implementations
// when connecting to the Brain. Note that this interface actually exposes direct
// write access to the events channel to allow adapters to deliver events
// synchronously and in deterministic order.
type EventRegistry interface {
	Channel() chan<- Event
	RegisterHandler(function interface{})
}

// brainRegistry implements the EventRegistry to connect a Brain with its Adapter.
type brainRegistry struct {
	*Brain
}

// Channel returns the events channel of the brain.
func (a brainRegistry) Channel() chan<- Event {
	return a.events
}

// NewBrain creates a new robot Brain. By default the Brain will use a Memory
// implementation that stores all keys and values directly in memory. You can
// change the memory implementation afterwards by simply assigning to
// Brain.Memory. If the passed logger is nil it will fallback to the
// zap.NewNop() logger. By default no timeout will be enforced on the event
// handlers.
func NewBrain(logger *zap.Logger) *Brain {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Brain{
		logger:   logger,
		memory:   newInMemory(),
		events:   make(chan Event, 10),
		handlers: make(map[reflect.Type][]eventHandler),
	}
}

// RegisterHandler registers a function to be executed when a specific event is
// fired. The function signature must comply with the following rules or the bot
// that uses this Brain will return an error on its next Bot.Run() call:
//
// Allowed function signatures:
//
//   // MyCustomEventStruct must be any struct but not a pointer to a struct.
//   func(MyCustomEventStruct)
//
//   // You can optionally accept a context as the first argument. It will
//   // receive the correct context of the Bot
//   func(context.Context, MyCustomEventStruct)
//
//   // You can optionally return a single error value. Returning any other type
//   // or returning more than one value will lead to an error. If the handler
//   // returns an error it will be logged.
//   func(MyCustomEventStruct) error
//
// The event that will be dispatched to the passed handler function corresponds
// directly to the accepted function argument. For instance if you want to emit
// and receive a custom event you can implement it like this:
//
//     type CustomEvent struct {}
//
//     b := NewBrain(nil)
//     b.RegisterHandler(func(
func (b *Brain) RegisterHandler(fun interface{}) {
	err := b.registerHandler(fun)
	if err != nil {
		caller := firstExternalCaller()
		err = errors.Wrap(err, caller)
		b.registrationErrs = append(b.registrationErrs, err)
	}
}

func (b *Brain) registerHandler(fun interface{}) error {
	handler := reflect.ValueOf(fun)
	handlerType := handler.Type()
	if handlerType.Kind() != reflect.Func {
		return errors.New("event handler is no function")
	}

	evtType, withContext, err := checkHandlerParams(handlerType)
	if err != nil {
		return err
	}

	returnsErr, err := checkHandlerReturnValues(handlerType)
	if err != nil {
		return err
	}

	b.logger.Debug("Registering new event handler",
		zap.Stringer("event_type", evtType),
	)

	handlerFun := newHandlerFunc(handler, withContext, returnsErr)
	b.handlers[evtType] = append(b.handlers[evtType], handlerFun)
	return nil
}

func (b *Brain) connectAdapter(a Adapter) {
	a.Register(brainRegistry{b})
}

// Emit sends the first argument as event to the brain from where it is
// dispatched to all registered handlers.
func (b *Brain) Emit(event interface{}, callbacks ...func(Event)) {
	go func() {
		b.events <- Event{Data: event, Callbacks: callbacks}
	}()
}

// HandleEvents starts the event handler loop of the Brain. This function blocks
// until the passed context is cancelled. If no handler timeout was configured
// the brain might block indefinitely even if the context is canceled but an
// event handler or callback is not respecting the context.
func (b *Brain) HandleEvents(ctx context.Context) {
	b.handleEvent(ctx, Event{Data: InitEvent{}})

	for {
		select {
		case evt := <-b.events:
			b.handleEvent(ctx, evt)

		case <-ctx.Done():
			b.handleEvent(ctx, Event{Data: ShutdownEvent{}})
			return
		}
	}
}

// handleEvent receives an event and determines which handler it must be
// dispatched to using the reflect API. Additionally the function enforces any
// event handler timeouts (if configured) and runs any event callbacks.
func (b *Brain) handleEvent(ctx context.Context, evt Event) {
	event := reflect.ValueOf(evt.Data)
	typ := event.Type()
	b.logger.Debug("Handling new event",
		zap.Stringer("event_type", typ),
		zap.Int("handlers", len(b.handlers[typ])),
	)

	for _, handler := range b.handlers[typ] {
		err := b.executeEventHandler(ctx, handler, event)
		if err != nil {
			b.logger.Error("Event handler failed",
				// TODO: somehow log the name of the handler
				zap.Error(err),
			)
		}
	}

	// TODO: callbacks should also get a context
	// TODO: respect context even if callbacks don't
	for _, callback := range evt.Callbacks {
		callback(evt)
	}
}

func (b *Brain) executeEventHandler(ctx context.Context, handler eventHandler, event reflect.Value) error {
	if b.handlerTimeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, b.handlerTimeout)
		defer cancel()
	}

	done := make(chan error)
	go func() {
		done <- handler(ctx, event)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Set is a wrapper around the Brains Memory.Set function to allow concurrent
// access and emit the corresponding BrainMemoryEvent.
func (b *Brain) Set(key, value string) error {
	b.mu.Lock()
	b.logger.Debug("Writing data to memory", zap.String("key", key))
	err := b.memory.Set(key, value)
	b.mu.Unlock()

	b.Emit(BrainMemoryEvent{Operation: "set", Key: key, Value: value})
	return err
}

// Get is a wrapper around the Brains Memory.Get function to allow concurrent
// access and emit the corresponding BrainMemoryEvent.
func (b *Brain) Get(key string) (string, bool, error) {
	b.mu.RLock()
	b.logger.Debug("Retrieving data from memory", zap.String("key", key))
	value, ok, err := b.memory.Get(key)
	b.mu.RUnlock()

	b.Emit(BrainMemoryEvent{Operation: "get", Key: key, Value: value})
	return value, ok, err
}

// Delete is a wrapper around the Brains Memory.Delete function to allow
// concurrent access and emit the corresponding BrainMemoryEvent.
func (b *Brain) Delete(key string) (bool, error) {
	b.mu.Lock()
	b.logger.Debug("Deleting data from memory", zap.String("key", key))
	ok, err := b.memory.Delete(key)
	b.mu.Unlock()

	b.Emit(BrainMemoryEvent{Operation: "del", Key: key})
	return ok, err
}

// Memories is a wrapper around the Brains Memory.Memories function to allow
// concurrent access.
func (b *Brain) Memories() (map[string]string, error) {
	b.mu.RLock()
	data, err := b.memory.Memories()
	b.mu.RUnlock()

	return data, err
}

// Close is a wrapper around the Brains Memory.Close function to allow
// concurrent access.
func (b *Brain) Close() error {
	b.mu.Lock()
	b.logger.Debug("Shutting down brain")
	err := b.memory.Close()
	b.mu.Unlock()

	return err
}
