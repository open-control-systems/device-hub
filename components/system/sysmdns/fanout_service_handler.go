package sysmdns

import "github.com/open-control-systems/device-hub/components/system/syscore"

// FanoutServiceHandler notifies the underlying handlers about discovered mDNS service.
type FanoutServiceHandler struct {
	handlers []ServiceHandler
}

// HandleService handles mDNS service discovered over local network.
func (h *FanoutServiceHandler) HandleService(service *Service) error {
	for _, handler := range h.handlers {
		if err := handler.HandleService(service); err != nil {
			syscore.LogErr.Printf("failed to handle mDNS service: %v", err)
		}
	}

	return nil
}

// Add adds handler to be notified when mDNS service is discovered.
func (h *FanoutServiceHandler) Add(handler ServiceHandler) {
	h.handlers = append(h.handlers, handler)
}
