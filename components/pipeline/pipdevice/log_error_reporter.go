package pipdevice

import "github.com/open-control-systems/device-hub/components/core"

type logErrorReporter struct {
	uri  string
	desc string
}

func (r *logErrorReporter) ReportError(err error) {
	core.LogErr.Printf("failed to handle device data: uri=%s desc=%s err=%v\n",
		r.uri, r.desc, err)
}
