package joe

import (
	"fmt"
	"runtime"
	"strings"
)

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
