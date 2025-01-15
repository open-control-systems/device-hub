package sysnet

import (
	"context"
	"net"
)

// TCPResolver resolves network addresses for TCP connections.
type TCPResolver struct{}

// Resolve resolves TCP address.
func (*TCPResolver) Resolve(_ context.Context, address string) (net.Addr, error) {
	return net.ResolveTCPAddr("tcp", address)
}
