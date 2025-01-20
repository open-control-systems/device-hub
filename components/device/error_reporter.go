package device

// ErrorReporter reports errors.
type ErrorReporter interface {
	// ReportError reports errors received from the device.
	//
	// Parameters:
	//	- uri - device URI.
	//	- desc - human readable device description.
	//	- err - error received from the device.
	ReportError(uri string, desc string, err error)
}
