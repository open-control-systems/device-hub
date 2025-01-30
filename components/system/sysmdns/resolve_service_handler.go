package sysmdns

import (
	"net"
	"strings"

	"github.com/open-control-systems/device-hub/components/system/sysnet"
)

// ResolveServiceHandler notifies about resolving results over local network.
type ResolveServiceHandler struct {
	handler sysnet.ResolveHandler
}

// NewResolveServiceHandler is an initialization of ResolveServiceHandler.
func NewResolveServiceHandler(handler sysnet.ResolveHandler) *ResolveServiceHandler {
	return &ResolveServiceHandler{handler: handler}
}

// HandleService handles mDNS service discovered over local network.
func (h *ResolveServiceHandler) HandleService(service *Service) error {
	addrs := service.AddrsIPv4

	if len(addrs) == 1 {
		h.handler.HandleResolve(
			strings.TrimSuffix(service.Hostname, "."),
			&net.IPAddr{IP: addrs[0]},
		)
	}

	return nil
}
