package pipdevice

import "github.com/open-control-systems/device-hub/components/core"

// LogErrorReporter reports errors from the device to the log.
type LogErrorReporter struct {
}

// ReportError reports device error to the log.
func (*LogErrorReporter) ReportError(uri string, desc string, err error) {
	core.LogErr.Printf("failed to handle device data: uri=%s desc=%s err=%v\n", uri, desc, err)
}
