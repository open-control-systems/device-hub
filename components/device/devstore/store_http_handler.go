package devstore

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/open-control-systems/device-hub/components/http/htcore"
)

// StoreHTTPHandler allows to add/remove devices over HTTP API.
type StoreHTTPHandler struct {
	store Store
}

// NewStoreHTTPHandler is an initialization of StoreHTTPHandler.
//
// Parameters:
//   - store to add/remove devices.
func NewStoreHTTPHandler(store Store) *StoreHTTPHandler {
	return &StoreHTTPHandler{store: store}
}

// HandleAdd adds the device over HTTP API.
func (h *StoreHTTPHandler) HandleAdd(w http.ResponseWriter, r *http.Request) {
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

	typ := r.URL.Query().Get("type")
	if typ == "" {
		http.Error(w, "error: missed `type` query parameter", http.StatusBadRequest)

		return
	}

	if err := h.store.Add(uri, typ, desc); err != nil {
		http.Error(w, fmt.Sprintf("error: failed to add device with uri=%s: %v", uri, err),
			http.StatusBadRequest)

		return
	}

	htcore.WriteText(w, "OK")
}

// HandleRemove removes the device over HTTP API.
func (h *StoreHTTPHandler) HandleRemove(w http.ResponseWriter, r *http.Request) {
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
func (h *StoreHTTPHandler) HandleList(w http.ResponseWriter, r *http.Request) {
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
