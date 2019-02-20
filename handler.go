package joe

import (
	"regexp"

	"github.com/pkg/errors"
)

type RespondFunc func(Message) error

type responseHandler struct {
	regex *regexp.Regexp
	run   RespondFunc
}

func newHandler(expr string, fun RespondFunc) (responseHandler, error) {
	safeHandler := func(msg Message) (handlerErr error) {
		defer func() {
			if err := recover(); err != nil {
				handlerErr = errors.Errorf("handler panic: %v", err)
			}
		}()

		return fun(msg)
	}

	h := responseHandler{run: safeHandler}

	var err error
	h.regex, err = regexp.Compile(expr)
	if err != nil {
		return h, errors.Wrap(err, "invalid regular expression")
	}

	return h, nil
}
