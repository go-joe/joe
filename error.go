package joe

// Error is the error type used by Joe. This allows joe errors to be defined as
// constants following https://dave.cheney.net/2016/04/07/constant-errors.
type Error string

// Error implements the "error" interface of the standard library.
func (err Error) Error() string {
	return string(err)
}
