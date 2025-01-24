package sysmdns

import "github.com/open-control-systems/device-hub/components/core"

// FanoutServiceHandler notifies the underlying handlers about discovered mDNS service.
type FanoutServiceHandler struct {
	handlers []ServiceHandler
}

// HandleService handles mDNS service discovered over local network.
func (h *FanoutServiceHandler) HandleService(service Service) error {
	for _, handler := range h.handlers {
		if err := handler.HandleService(service); err != nil {
			core.LogErr.Printf("fanout-service-handler: failed to handle mDNS service: %v\n",
				err)
		}
	}

	return nil
}

// Add adds handler to be notified when mDNS service is discovered.
func (h *FanoutServiceHandler) Add(handler ServiceHandler) {
	h.handlers = append(h.handlers, handler)
}
