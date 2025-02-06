package hthandler

import (
	"fmt"
	"net/http"

	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// CrashHandler to handle panics during HTTP request handling.
type CrashHandler struct {
	handler http.Handler
}

// NewCrashHandler is an initialization of CrashHandler.
func NewCrashHandler(handler http.Handler) *CrashHandler {
	return &CrashHandler{
		handler: handler,
	}
}

// ServeHTTP wraps the underlying HTTP handler with panic-recovery mechanism.
func (h *CrashHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			if e == http.ErrAbortHandler {
				panic(e)
			}

			syscore.LogCrash(e)

			http.Error(w, fmt.Sprintf("crash: %#v", e), http.StatusInternalServerError)
		}
	}()

	h.handler.ServeHTTP(w, r)
}
