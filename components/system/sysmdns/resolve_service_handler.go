package sysmdns

import (
	"fmt"
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
func (h *ResolveServiceHandler) HandleService(service Service) error {
	addrs := service.Addrs()
	if len(addrs) < 1 {
		return fmt.Errorf("ignore service: instance=%s service=%s hostname=%s:"+
			" IP address not found",
			service.Instance(), service.Name(), service.Hostname())
	}

	hostname := strings.TrimSuffix(service.Hostname(), ".")
	addr := &net.IPAddr{IP: addrs[0]}

	h.handler.HandleResolve(hostname, addr)

	return nil
}
