package htcore

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// SystemClockHandler handles the UNIX time configuration over HTTP.
type SystemClockHandler struct {
	clock      syscore.SystemClock
	startPoint time.Time
}

// NewSystemClockHandler creates an HTTP handler for the UNIX time configuration.
func NewSystemClockHandler(
	clock syscore.SystemClock,
	startPoint time.Time,
) *SystemClockHandler {
	return &SystemClockHandler{
		clock:      clock,
		startPoint: startPoint,
	}
}

// ServeHTTP implements an HTTP endpoint logic.
func (h *SystemClockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "error: unsupported method", http.StatusMethodNotAllowed)

		return
	}

	response := ""

	str := r.URL.Query().Get("value")
	if str == "" {
		timestamp, err := h.clock.GetTimestamp()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get UNIX time: %v", err),
				http.StatusInternalServerError)

			return
		}

		if timestamp < h.startPoint.Unix() {
			timestamp = -1
		}

		response = strconv.FormatInt(timestamp, 10)
	} else {
		timestamp, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		if err := h.clock.SetTimestamp(timestamp); err != nil {
			http.Error(w, fmt.Sprintf("failed to set UNIX time: %v", err),
				http.StatusInternalServerError)

			return
		}

		response = "OK"
	}

	WriteText(w, response)
}
