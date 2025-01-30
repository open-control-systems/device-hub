package sysmdns

import (
	"net"
	"testing"

	"github.com/open-control-systems/device-hub/components/status"
	"github.com/stretchr/testify/require"
)

type testResolveServiceHandlerResolveHandler struct {
	host string
	addr net.Addr
}

func (h *testResolveServiceHandlerResolveHandler) HandleResolve(host string, addr net.Addr) {
	h.host = host
	h.addr = addr
}

func TestResolveServiceHandlerIPv4(t *testing.T) {
	resolveHandler := &testResolveServiceHandlerResolveHandler{}
	serviceHandler := NewResolveServiceHandler(resolveHandler)

	hostname := "foo"
	port := 8081
	addr := net.IPAddr{IP: net.IPv4(192, 168, 0, 10)}

	require.Nil(t, serviceHandler.HandleService(&Service{
		Hostname:  hostname,
		Port:      port,
		AddrsIPv4: []net.IP{addr.IP},
	}))
	require.Equal(t, hostname, resolveHandler.host)
	require.Equal(t, addr.String(), resolveHandler.addr.String())
}

func TestResolveServiceHandlerIPv4Many(t *testing.T) {
	resolveHandler := &testResolveServiceHandlerResolveHandler{}
	serviceHandler := NewResolveServiceHandler(resolveHandler)

	hostname := "foo"
	port := 8081

	addr1 := net.IPAddr{IP: net.IPv4(192, 168, 0, 10)}
	addr2 := net.IPAddr{IP: net.IPv4(192, 168, 0, 11)}
	require.NotEqual(t, addr1.String(), addr2.String())

	require.Equal(t, status.StatusNotSupported, serviceHandler.HandleService(&Service{
		Hostname:  hostname,
		Port:      port,
		AddrsIPv4: []net.IP{addr1.IP, addr2.IP},
	}))
	require.Empty(t, resolveHandler.host)
	require.Nil(t, resolveHandler.addr)
}
