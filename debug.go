package joe

import (
	"fmt"
	"runtime"
)

func firstExternalCaller() {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	callers := pcs[0:n]

	frames := runtime.CallersFrames(callers)
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		fmt.Printf("%+v\n", frame)
	}
}
