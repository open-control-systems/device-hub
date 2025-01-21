package core

// FuncCloser is a function type that implements the Closer interface.
type FuncCloser func() error

// Close calls the function itself to fulfill the Closer interface.
func (f FuncCloser) Close() error {
	return f()
}
