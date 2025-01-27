package devcore

import "sync"

// IDHolder holds the device unique identifier.
type IDHolder struct {
	handler DataHandler

	mu sync.RWMutex
	id string
}

// NewIDHolder is an initialization of IDHolder.
//
// Parameters:
//   - handler to propagate the actual calls for the device data handling.
func NewIDHolder(handler DataHandler) *IDHolder {
	return &IDHolder{
		handler: handler,
	}
}

// Get returns the unique device identifier.
//
// Remarks:
//   - Can be used by multiple goroutines.
func (h *IDHolder) Get() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.id
}

// HandleRegistration updates the device ID.
//
// Remarks:
//   - Should be used by a single goroutine.
//   - Device ID is changed very rarely, rw-lock is used to reduce contention.
func (h *IDHolder) HandleRegistration(deviceID string, js JSON) error {
	h.mu.RLock()
	id := h.id
	h.mu.RUnlock()

	if id != deviceID {
		h.mu.Lock()
		h.id = deviceID
		h.mu.Unlock()
	}

	return h.handler.HandleRegistration(deviceID, js)
}

// HandleTelemetry propagates call to the underlying data handler.
func (h *IDHolder) HandleTelemetry(deviceID string, js JSON) error {
	return h.handler.HandleTelemetry(deviceID, js)
}
