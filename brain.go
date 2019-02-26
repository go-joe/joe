package joe

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Brain struct {
	mu     sync.RWMutex
	memory Memory
	logger *zap.Logger

	events         chan event
	handlers       map[reflect.Type][]eventHandler
	handlerTimeout time.Duration // zero means no timeout

	registrationErrs []error
}

type event struct {
	Data      interface{}
	callbacks []func(event)
}

type eventHandler func(context.Context, reflect.Value) error

func NewBrain(logger *zap.Logger) *Brain {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Brain{
		logger:   logger,
		memory:   newInMemory(),
		events:   make(chan event, 10),
		handlers: make(map[reflect.Type][]eventHandler),
	}
}

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

func (b *Brain) Emit(eventData interface{}, callbacks ...func(event)) {
	go func() {
		b.events <- event{Data: eventData, callbacks: callbacks}
	}()
}

func (b *Brain) HandleEvents(ctx context.Context) {
	for {
		select {
		case evt := <-b.events:
			b.handleEvent(ctx, evt)

		case <-ctx.Done():
			b.handleEvent(ctx, event{Data: ShutdownEvent{}})
			return
		}
	}
}

func (b *Brain) handleEvent(ctx context.Context, evt event) {
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
	for _, callback := range evt.callbacks {
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

func (b *Brain) Set(key, value string) error {
	b.mu.Lock()
	b.logger.Debug("Writing data to memory", zap.String("key", key))
	err := b.memory.Set(key, value)
	b.mu.Unlock()

	b.Emit(BrainMemoryEvent{Operation: "set", Key: key, Value: value})
	return err
}

func (b *Brain) Get(key string) (string, bool, error) {
	b.mu.RLock()
	b.logger.Debug("Retrieving data from memory", zap.String("key", key))
	value, ok, err := b.memory.Get(key)
	b.mu.RUnlock()

	b.Emit(BrainMemoryEvent{Operation: "get", Key: key, Value: value})
	return value, ok, err
}

func (b *Brain) Delete(key string) (bool, error) {
	b.mu.Lock()
	b.logger.Debug("Deleting data from memory", zap.String("key", key))
	ok, err := b.memory.Delete(key)
	b.mu.Unlock()

	b.Emit(BrainMemoryEvent{Operation: "del", Key: key})
	return ok, err
}

func (b *Brain) Memories() (map[string]string, error) {
	b.mu.RLock()
	data, err := b.memory.Memories()
	b.mu.RUnlock()

	return data, err
}

func (b *Brain) Close() error {
	b.mu.Lock()
	b.logger.Debug("Shutting down brain")
	err := b.memory.Close()
	b.mu.Unlock()

	return err
}
