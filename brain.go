package joe

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// The Brain contains the core logic of a Bot by implementing an event handling
// system that dispatches events to all registered event handlers.
type Brain struct {
	logger *zap.Logger

	eventsInput chan Event // input for any new events, the Brain ensures that callers never block when writing to it
	eventsLoop  chan Event // used in Brain.HandleEvents() to actually process the events
	shutdown    chan shutdownRequest

	mu             sync.RWMutex // mu protects concurrent access to the handlers
	handlers       map[reflect.Type][]eventHandler
	handlerTimeout time.Duration // zero means no timeout, defaults to one minute

	registrationErrs []error // any errors that occurred during setup (e.g. in Bot.RegisterHandler)
	handlingEvents   int32   // accessed atomically (non-zero means the event handler was started)
	closed           int32   // accessed atomically (non-zero means the brain was shutdown already)
}

// An Event represents a concrete event type and optional callbacks that are
// triggered when the event was processed by all registered handlers.
type Event struct {
	Data       interface{}
	Callbacks  []func(Event)
	AbortEarly bool
}

// The shutdownRequest type is used when signaling shutdown information between
// Brain.Shutdown() and the Brain.HandleEvents loop.
type shutdownRequest struct {
	ctx      context.Context
	callback chan bool
}

// An eventHandler is a function that takes a context and the reflected value
// of a concrete event type.
type eventHandler func(context.Context, reflect.Value) error

// ctxKey is used to pass meta information to event handlers via the context.
type ctxKey string

// ctxKeyEvent is the context key under which we can lookup the internal *Event
// instance in a handler.
const ctxKeyEvent ctxKey = "event"

// FinishEventContent can be called from within your event handler functions
// to indicate that the Brain should not execute any other handlers after the
// calling handler has returned.
func FinishEventContent(ctx context.Context) {
	evt, _ := ctx.Value(ctxKeyEvent).(*Event)
	if evt != nil {
		evt.AbortEarly = true
	}
}

// NewBrain creates a new robot Brain. If the passed logger is nil it will
// fallback to the zap.NewNop() logger.
func NewBrain(logger *zap.Logger) *Brain {
	if logger == nil {
		logger = zap.NewNop()
	}

	b := &Brain{
		logger:         logger,
		eventsInput:    make(chan Event),
		eventsLoop:     make(chan Event),
		shutdown:       make(chan shutdownRequest),
		handlers:       make(map[reflect.Type][]eventHandler),
		handlerTimeout: time.Minute,
	}

	b.consumeEvents()

	return b
}

func (b *Brain) isHandlingEvents() bool {
	return atomic.LoadInt32(&b.handlingEvents) == 1
}

func (b *Brain) isClosed() bool {
	return atomic.LoadInt32(&b.closed) == 1
}

// RegisterHandler registers a function to be executed when a specific event is
// fired. The function signature must comply with the following rules or the bot
// that uses this Brain will return an error on its next Bot.Run() call:
//
// Allowed function signatures:
//
//   // AnyType can be any scalar, struct or interface type as long as it is not
//   // a pointer.
//   func(AnyType)
//
//   // You can optionally accept a context as the first argument. The context
//   // is used to signal handler timeouts or when the bot is shutting down.
//   func(context.Context, AnyType)
//
//   // You can optionally return a single error value. Returning any other type
//   // or returning more than one value is not possible. If the handler
//   // returns an error it will be logged.
//   func(AnyType) error
//
//   // Event handlers can also accept an interface in which case they will be
//   // be called for all events which implement the interface. Consequently,
//   // you can register a function which accepts the empty interface which will
//   // will receive all emitted events. Such event handlers can optionally also
//   // accept a context and/or return an error like other handlers.
//   func(context.Context, interface{}) error
//
// The event, that will be dispatched to the passed handler function, corresponds
// directly to the accepted function argument. For instance if you want to emit
// and receive a custom event you can implement it like this:
//
//     type CustomEvent struct {}
//
//     b := NewBrain(nil)
//     b.RegisterHandler(func(evt CustomEvent) {
//         …
//     })
//
// If multiple handlers are registered for the same event type, then they are
// all executed in the order in which they have been registered.
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

	b.mu.Lock()
	b.handlers[evtType] = append(b.handlers[evtType], handlerFun)
	b.mu.Unlock()

	return nil
}

// Emit sends the first argument as event to the brain from where it is
// dispatched to all registered handlers. The events are dispatched
// asynchronously but in the same order in which they are send to this function.
// Emit does not block until the event is delivered to the registered event
// handlers. If you want to wait until all handlers have processed the event you
// can pass one or more callback functions that will be executed when all
// handlers finished execution of this event.
func (b *Brain) Emit(event interface{}, callbacks ...func(Event)) {
	if b.isClosed() {
		b.logger.Debug(
			"Ignoring new event because brain is currently shutting down or is already closed",
			zap.String("type", fmt.Sprintf("%T", event)),
		)
		return
	}

	b.eventsInput <- Event{Data: event, Callbacks: callbacks}
}

// HandleEvents starts the event handling loop of the Brain.
// This function blocks until Brain.Shutdown() is called and returned.
func (b *Brain) HandleEvents() {
	if b.isClosed() {
		b.logger.Error("HandleEvents failed because bot is already closed")
		return
	}

	ctx := context.Background()
	var shutdown shutdownRequest // set when Brain.Shutdown() is called

	atomic.StoreInt32(&b.handlingEvents, 1)
	b.handleEvent(ctx, Event{Data: InitEvent{}})

	for {
		select {
		case evt, ok := <-b.eventsLoop:
			if !ok {
				// Brain.consumeEvents() is done processing all remaining events
				// and we can now safely shutdown the event handler, knowing that
				// all pending events have been processed.
				b.handleEvent(ctx, Event{Data: ShutdownEvent{}})
				shutdown.callback <- true
				return
			}

			b.handleEvent(ctx, evt)

		case shutdown = <-b.shutdown:
			// The Brain is shutting down. We have to close the input channel so
			// we doe no longer accept new events and only process the remaining
			// pending events. When the goroutine of Brain.consumeEvents() is
			// done it will close the events loop channel and the case above will
			// use the shutdown callback and return from this function.
			ctx = shutdown.ctx
			close(b.eventsInput)
			atomic.StoreInt32(&b.handlingEvents, 0)
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
func (b *Brain) handleEvent(ctx context.Context, evt Event) {
	event := reflect.ValueOf(evt.Data)
	typ := event.Type()
	handlers := b.determineHandlers(typ)

	b.logger.Debug("Handling new event",
		zap.Stringer("event_type", typ),
		zap.Int("handlers", len(handlers)),
	)

	ctx = context.WithValue(ctx, ctxKeyEvent, &evt)

	for _, handler := range handlers {
		err := b.executeEventHandler(ctx, handler, event)
		if err != nil {
			b.logger.Error("Event handler failed",
				// TODO: somehow log the name of the handler
				zap.Error(err),
			)
		}

		if evt.AbortEarly {
			// Abort handler execution early instead of running any more
			// handlers. The event state may have been changed by a handler, e.g.
			// using the FinishEventContent(…) function.
			break
		}
	}

	for _, callback := range evt.Callbacks {
		callback(evt)
	}
}

func (b *Brain) determineHandlers(evtType reflect.Type) []eventHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var handlers []eventHandler
	for handlerType, hh := range b.handlers {
		if handlerType == evtType {
			handlers = append(handlers, hh...)
		}

		if handlerType.Kind() == reflect.Interface && evtType.Implements(handlerType) {
			handlers = append(handlers, hh...)
		}
	}

	return handlers
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

// Shutdown stops the event handler loop of the Brain and waits until all pending
// events have been processed. After the brain is shutdown, it will no longer
// accept new events. The passed context can be used to stop waiting for any
// pending events or handlers and instead exit immediately (e.g. after a timeout
// or a second SIGTERM).
func (b *Brain) Shutdown(ctx context.Context) {
	closing := atomic.CompareAndSwapInt32(&b.closed, 0, 1)
	if !closing {
		// brain is already shutting down
		return
	}

	if !b.isHandlingEvents() {
		// If the event handler loop is not running we must close the inputs
		// channel from here and drain all pending requests in order to make
		// b.consumeEvents() exit.
		close(b.eventsInput)
		for {
			select {
			case _, ok := <-b.eventsLoop:
				if !ok {
					// The eventsLoop channel is closed in b.consumeEvents after
					// all pending messages have been written to it.
					return
				}
			case <-ctx.Done():
				// shutdown context is expired so we return without waiting for
				// any pending events.
				return
			}
		}
	}

	// If we got here then the event handler loop is running and we delegate
	// proper cleanup and processing of pending messages over there.
	req := shutdownRequest{
		ctx:      ctx,
		callback: make(chan bool),
	}

	b.shutdown <- req
	<-req.callback
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

	if evtType.Kind() == reflect.Ptr {
		err = errors.New("event handler argument cannot be a pointer")
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
