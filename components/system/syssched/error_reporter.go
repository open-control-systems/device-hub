package syssched

// ErrorReporter reports errors.
type ErrorReporter interface {
	// ReportError reports errors received from the task.
	ReportError(err error)
}
