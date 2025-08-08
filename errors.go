package memoryguard

const (
	// LimitZeroError is returned by Limit(int64) when the passed variable is <= 0.
	LimitZeroError = Error("please call Limit(int64) with a value greater than zero")
	// LimitNilProcessError is returned by Limit(int64) when the referenced *os.Process is nil.
	LimitNilProcessError = Error("a Process has not been created and assigned, or is nil")
	// LimitOnceError is returned by Limit(int64) if it has been called without error previously.
	LimitOnceError = Error("Limit(int64) already called once")
)

// Error is an error type
type Error string

// Error returns the stringified version of Error
func (e Error) Error() string {
	return string(e)
}
