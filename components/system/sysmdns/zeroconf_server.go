package sysmdns

import (
	"net"

	"github.com/open-control-systems/zeroconf"
)

// ZeroconfServer registers new mDNS services.
type ZeroconfServer struct {
	services []*Service
	ifaces   []net.Interface
	servers  []*zeroconf.Server
}

// NewZeroconfServer is an initialization of ZeroconfServer.
func NewZeroconfServer(services []*Service, ifaces []net.Interface) *ZeroconfServer {
	return &ZeroconfServer{
		services: services,
		ifaces:   ifaces,
	}
}

// Start starts all registered mDNS services.
func (s *ZeroconfServer) Start() error {
	for _, service := range s.services {
		server, err := zeroconf.RegisterProxy(
			service.Instance, service.Name, "local",
			service.Port, service.Hostname, nil,
			service.TxtRecords, s.ifaces,
		)
		if err != nil {
			return err
		}

		s.servers = append(s.servers, server)
	}

	return nil
}

// Stop cleans up all allocated resources.
func (s *ZeroconfServer) Stop() error {
	for _, server := range s.servers {
		server.Shutdown()
	}

	return nil
}
