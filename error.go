package joe

// Error is the error type used by Joe. This allows joe errors to be defined as
// constants following https://dave.cheney.net/2016/04/07/constant-errors.
type Error string

// Error implements the "error" interface of the standard library.
func (err Error) Error() string {
	return string(err)
}

// ErrNotImplemented is returned if the user tries to use a feature that is not
// implemented on the corresponding components (e.g. the Adapter). For instance,
// not all Adapter implementations may support emoji reactions and trying to
// attach a reaction to a message might return this error.
const ErrNotImplemented = Error("not implemented")
