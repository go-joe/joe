package joe

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

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

	eventsInput chan Event // input for any new events, the Brain ensures that callers never block when writing to it
	eventsLoop  chan Event // used in Brain.HandleEvents() to actually process the events
	shutdown    chan chan bool

	handlers map[reflect.Type][]eventHandler

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
	return a.eventsInput
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

	b := &Brain{
		logger:      logger,
		memory:      newInMemory(),
		eventsInput: make(chan Event),
		eventsLoop:  make(chan Event),
		shutdown:    make(chan chan bool),
		handlers:    make(map[reflect.Type][]eventHandler),
	}

	b.consumeEvents()

	return b
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
// dispatched to all registered handlers. This function blocks until the event
// handler was started via Brain.HandleEvents(…). When the event handler is
// running it will return immediately.
func (b *Brain) Emit(event interface{}, callbacks ...func(Event)) {
	// TODO: do not panic if Brain is already shutting down
	b.eventsInput <- Event{Data: event, Callbacks: callbacks}
}

// HandleEvents starts the event handler loop of the Brain. This function blocks
// until Brain.Shutdown() is canceled. If no handler timeout was configured
// the brain might block indefinitely even if the brain is shutting down but an
// event handler or callback is unresponsive.
func (b *Brain) HandleEvents() {
	b.handleEvent(Event{Data: InitEvent{}})

	var shutdownCallback chan bool // set when Brain.Shutdown() is called

	for {
		select {
		case evt, ok := <-b.eventsLoop:
			if !ok {
				// Brain.consumeEvents() is done processing all remaining events
				// and we can now safely shutdown the event handler, knowing that
				// all pending events have been processed.
				b.handleEvent(Event{Data: ShutdownEvent{}})
				shutdownCallback <- true
				return
			}

			b.handleEvent(evt)

		case shutdownCallback = <-b.shutdown:
			// The Brain is shutting down. We have to close the input channel so
			// we doe no longer accept new events and only process the remaining
			// pending events. When the goroutine of Brain.consumeEvents() is
			// done it will close the events loop channel and the case above will
			// use the shutdown callback and return from this function.
			close(b.eventsInput)
		}
	}
}

// consumeEvents continuously reads events from b.eventsInput in a new goroutine
// so emitting an event never blocks on the caller. All events will be returned
// in the result channel of this function in the same order in which they have
// been inserted into b.events. In this sense this function provides an events
// channel with "infinite" capacity. The spawned goroutine stops when the
// b.eventsInput channel is closed.
func (b *Brain) consumeEvents() {
	var queue []Event
	b.eventsLoop = make(chan Event)

	outChan := func() chan Event {
		if len(queue) == 0 {
			// In case the queue is empty we return a nil channel to disable the
			// corresponding select case in the goroutine below.
			return nil
		}

		return b.eventsLoop
	}

	nextEvt := func() Event {
		if len(queue) == 0 {
			// Prevent index out of bounds if there is no next event. Note that
			// this event is actually never received because the outChan()
			// function above will return "nil" in this case which disables the
			// corresponding select case.
			return Event{}
		}

		return queue[0]
	}

	go func() {
		for {
			select {
			case evt, ok := <-b.eventsInput:
				if !ok {
					// Events input channel was closed because Brain is shutting
					// down. Emit all pending events from the queue and then close
					// the events loop channel so Brain.HandleEvents() can exit.
					for _, evt := range queue {
						b.eventsLoop <- evt
					}
					close(b.eventsLoop)
					return
				}

				queue = append(queue, evt)
			case outChan() <- nextEvt(): // disabled if len(queue) == 0
				queue = queue[1:]
			}
		}
	}()
}

// handleEvent receives an event and dispatches it to all registered handlers
// using the reflect API. When all applicable handlers are called (maybe none)
// the function runs all event callbacks.
func (b *Brain) handleEvent(evt Event) {
	event := reflect.ValueOf(evt.Data)
	typ := event.Type()
	b.logger.Debug("Handling new event",
		zap.Stringer("event_type", typ),
		zap.Int("handlers", len(b.handlers[typ])),
	)

	ctx := context.TODO() // TODO, what do we want here?
	for _, handler := range b.handlers[typ] {
		err := handler(ctx, event)
		if err != nil {
			b.logger.Error("Event handler failed",
				// TODO: somehow log the name of the handler
				zap.Error(err),
			)
		}
	}

	for _, callback := range evt.Callbacks {
		callback(evt)
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

func (b *Brain) Shutdown() {
	// TODO: do not block if Brain was not yet started (unit test)
	callback := make(chan bool)
	b.shutdown <- callback
	<-callback
}

func checkHandlerParams(handlerFunc reflect.Type) (evtType reflect.Type, withContext bool, err error) {
	numParams := handlerFunc.NumIn()
	if numParams == 0 || numParams > 2 {
		err = errors.New("event handler needs one or two arguments")
		return
	}

	evtType = handlerFunc.In(numParams - 1) // last argument must be the event
	withContext = numParams == 2

	if withContext {
		contextInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
		if handlerFunc.In(1).Implements(contextInterface) {
			err = errors.New("event handler context must be the first argument")
			return
		}
		if !handlerFunc.In(0).Implements(contextInterface) {
			err = errors.New("event handler has two arguments but the first is not a context.Context")
			return
		}
	}

	switch evtType.Kind() {
	case reflect.Struct:
		// ok cool, move on
	case reflect.Ptr:
		err = errors.New("event handler argument must be a struct and not a pointer")
		return
	default:
		err = errors.New("event handler argument must be a struct")
		return
	}

	return evtType, withContext, nil
}

func checkHandlerReturnValues(handlerFunc reflect.Type) (returnsError bool, err error) {
	switch handlerFunc.NumOut() {
	case 0:
		return false, nil
	case 1:
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !handlerFunc.Out(0).Implements(errorInterface) {
			err = errors.New("if the event handler has a return value it must implement the error interface")
			return
		}
		return true, nil
	default:
		return false, errors.Errorf("event handler has more than one return value")
	}
}

func newHandlerFunc(handler reflect.Value, withContext, returnsErr bool) eventHandler {
	return func(ctx context.Context, evt reflect.Value) (handlerErr error) {
		defer func() {
			if err := recover(); err != nil {
				handlerErr = errors.Errorf("handler panic: %v", err)
			}
		}()

		var args []reflect.Value
		if withContext {
			args = []reflect.Value{
				reflect.ValueOf(ctx),
				evt,
			}
		} else {
			args = []reflect.Value{evt}
		}

		results := handler.Call(args)
		if returnsErr && !results[0].IsNil() {
			return results[0].Interface().(error)
		}

		return nil
	}
}

func firstExternalCaller() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	callers := pcs[0:n]

	frames := runtime.CallersFrames(callers)
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if !strings.HasPrefix(frame.Function, "github.com/go-joe/joe.") {
			return fmt.Sprintf("%s:%d", frame.File, frame.Line)
		}
	}

	return "unknown caller"
}
