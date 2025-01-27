package syssched

// ErrorHandler handles errors.
type ErrorHandler interface {
	// HandleError handles error.
	HandleError(err error)
}
