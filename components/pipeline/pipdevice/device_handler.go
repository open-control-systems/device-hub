package pipdevice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/open-control-systems/device-hub/components/http/htcore"
)

// DeviceHandler allows to add/remove devices over HTTP API.
type DeviceHandler struct {
	store *PipelineStore
}

// NewDeviceHandler is an initialization of DeviceHandler.
//
// Parameters:
//   - store to add/remove devices.
func NewDeviceHandler(store *PipelineStore) *DeviceHandler {
	return &DeviceHandler{store: store}
}

// HandleAdd adds the device over HTTP API.
func (h *DeviceHandler) HandleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "error: unsupported method", http.StatusMethodNotAllowed)

		return
	}

	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "error: missed `uri` query parameter", http.StatusBadRequest)

		return
	}

	desc := r.URL.Query().Get("desc")
	if desc == "" {
		http.Error(w, "error: missed `desc` query parameter", http.StatusBadRequest)

		return
	}

	if err := h.store.Add(uri, desc); err != nil {
		http.Error(w, fmt.Sprintf("error: failed to add device with uri=%s: %v", uri, err),
			http.StatusBadRequest)

		return
	}

	htcore.WriteText(w, "OK")
}

// HandleRemove removes the device over HTTP API.
func (h *DeviceHandler) HandleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "error: unsupported method", http.StatusMethodNotAllowed)

		return
	}

	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "error: missed `uri` query parameter", http.StatusBadRequest)

		return
	}

	if err := h.store.Remove(uri); err != nil {
		http.Error(w, fmt.Sprintf("error: failed to remove device with uri=%s: %v", uri, err),
			http.StatusBadRequest)

		return
	}

	htcore.WriteText(w, "OK")
}

// HandleList returns the description of all added devices.
func (h *DeviceHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "error: unsupported method", http.StatusMethodNotAllowed)

		return
	}

	buf, err := json.Marshal(h.store.GetDesc())
	if err != nil {
		http.Error(w, fmt.Sprintf("error: failed to format JSON: %v", err),
			http.StatusInternalServerError)

		return
	}

	htcore.WriteJSON(w, buf)
}
