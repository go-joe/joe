package joe

import (
	"context"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type EventProcessor struct {
	input          chan Event
	logger         *zap.Logger
	handlerTimeout time.Duration // zero means no timeout
	handlers       map[reflect.Type][]eventHandler
}

type Event struct {
	Data      interface{}
	callbacks []func(Event)
}

type eventHandler func(context.Context, reflect.Value) error

func NewEventProcessor(logger *zap.Logger, handlerTimeout time.Duration) *EventProcessor {
	return &EventProcessor{
		logger:         logger.Named("events"),
		input:          make(chan Event, 10),
		handlers:       make(map[reflect.Type][]eventHandler),
		handlerTimeout: handlerTimeout,
	}
}

func (p *EventProcessor) RegisterHandler(fun interface{}) {
	logErr := func(err error, fields ...zapcore.Field) {
		p.logger.Error("Failed to register handler: "+err.Error(), fields...)
	}

	handler := reflect.ValueOf(fun)
	handlerType := handler.Type()
	if handlerType.Kind() != reflect.Func {
		logErr(errors.New("event handler is no function"))
		return
	}

	evtType, withContext, err := p.checkHandlerParams(handlerType)
	if err != nil {
		logErr(err)
		return
	}

	returnsErr, err := p.checkHandlerReturnValues(handlerType)
	if err != nil {
		logErr(err)
		return
	}

	p.logger.Debug("Registering new event handler",
		zap.String("event_type", evtType.Name()),
	)

	handlerFun := p.newHandlerFunc(handler, withContext, returnsErr)
	p.handlers[evtType] = append(p.handlers[evtType], handlerFun)
}

func (*EventProcessor) checkHandlerParams(handlerFunc reflect.Type) (evtType reflect.Type, withContext bool, err error) {
	numParams := handlerFunc.NumIn()
	if numParams == 0 || numParams > 2 {
		err = errors.New("event handler function needs one or two arguments")
		return
	}

	evtType = handlerFunc.In(numParams - 1) // last argument must be the event
	withContext = numParams == 2

	if evtType.Kind() != reflect.Struct {
		err = errors.New("last event handler function argument must be a struct")
		return
	}

	if withContext {
		contextInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
		if !handlerFunc.In(0).Implements(contextInterface) {
			err = errors.New("event handler function argument 1 is not a context.Context")
			return
		}
	}

	return evtType, withContext, nil
}

func (*EventProcessor) checkHandlerReturnValues(handlerFunc reflect.Type) (returnsError bool, err error) {
	switch handlerFunc.NumOut() {
	case 0:
		return false, nil
	case 1:
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !handlerFunc.Out(0).Implements(errorInterface) {
			err = errors.New("event handler function return value must implement the error interface")
			return
		}
		return true, nil
	default:
		return false, errors.Errorf("event handler function has more than one return value")
	}
}

func (*EventProcessor) newHandlerFunc(handler reflect.Value, withContext, returnsErr bool) eventHandler {
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

func (p *EventProcessor) Process(ctx context.Context) {
	for {
		select {
		case evt := <-p.input:
			p.handle(ctx, evt)

		case <-ctx.Done():
			p.handle(ctx, Event{Data: ShutdownEvent{}})
			return
		}
	}
}

func (p *EventProcessor) handle(ctx context.Context, evt Event) {
	event := reflect.ValueOf(evt.Data)
	typ := event.Type()
	p.logger.Debug("Handling new event",
		zap.String("event_type", typ.Name()),
		zap.Int("handlers", len(p.handlers[typ])),
	)

	for _, handler := range p.handlers[typ] {
		err := p.executeHandler(ctx, handler, event)
		if err != nil {
			p.logger.Error("Event handler failed",
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

func (p *EventProcessor) executeHandler(ctx context.Context, handler eventHandler, event reflect.Value) error {
	if p.handlerTimeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, p.handlerTimeout)
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

func (p *EventProcessor) Emit(event interface{}, callbacks ...func(Event)) {
	go func() {
		p.input <- Event{Data: event, callbacks: callbacks}
	}()
}
