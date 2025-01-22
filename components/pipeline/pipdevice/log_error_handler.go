package pipdevice

import "github.com/open-control-systems/device-hub/components/core"

type logErrorHandler struct {
	uri  string
	desc string
}

func (h *logErrorHandler) HandleError(err error) {
	core.LogErr.Printf("device-error-handler: failed to handle device data:"+
		" uri=%s desc=%s err=%v\n", h.uri, h.desc, err)
}
