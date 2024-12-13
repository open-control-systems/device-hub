package status

import "errors"

var (
	// StatusError indicates a failure of an operation.
	StatusError = errors.New("operation failed")
)
