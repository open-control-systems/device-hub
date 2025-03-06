package devstore

import "github.com/open-control-systems/device-hub/components/system/syscore"

type logErrorHandler struct {
	uri  string
	typ  string
	desc string
}

func (h *logErrorHandler) HandleError(err error) {
	syscore.LogErr.Printf("failed to handle device data: uri=%s type=%s desc=%s err=%v",
		h.uri, h.typ, h.desc, err)
}
