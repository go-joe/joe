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
	h := responseHandler{run: fun}

	var err error
	h.regex, err = regexp.Compile(expr)
	if err != nil {
		return h, errors.Wrap(err, "invalid regular expression")
	}

	return h, nil
}
