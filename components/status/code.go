package status

import "errors"

var (
	// StatusError indicates a failure of an operation.
	StatusError = errors.New("operation failed")

	// StatusInvalidState indicates that an operation can't be performed due to invalid state.
	StatusInvalidState = errors.New("invalid state")

	// StatusNotSupported indicates that an operation isn't supported.
	StatusNotSupported = errors.New("not implemented")
)
