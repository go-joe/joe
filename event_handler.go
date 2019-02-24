package joe

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

func checkHandlerParams(handlerFunc reflect.Type) (evtType reflect.Type, withContext bool, err error) {
	numParams := handlerFunc.NumIn()
	if numParams == 0 || numParams > 2 {
		err = errors.New("event handler needs one or two arguments")
		return
	}

	evtType = handlerFunc.In(numParams - 1) // last argument must be the event
	withContext = numParams == 2

	if evtType.Kind() != reflect.Struct {
		err = errors.New("event handler argument must be a struct")
		return
	}

	if withContext {
		contextInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
		if !handlerFunc.In(0).Implements(contextInterface) {
			err = errors.New("event handler has 2 arguments but the first is not a context.Context")
			return
		}
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
			err = errors.New("if the event handler has a return value i must implement the error interface")
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
