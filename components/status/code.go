package status

import "errors"

var (
	// StatusError indicates a failure of an operation.
	StatusError = errors.New("operation failed")

	// StatusInvalidState indicates that an operation can't be performed due to invalid state.
	StatusInvalidState = errors.New("invalid state")

	// StatusNotSupported indicates that an operation isn't supported.
	StatusNotSupported = errors.New("not supported")

	// StatusNoData indicates that there is no enough data to perform an operation.
	StatusNoData = errors.New("no data")

	// StatusTimeout indicates that an operation was not performed within the timeout.
	StatusTimeout = errors.New("timeout")
)
