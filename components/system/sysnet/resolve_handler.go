package sysnet

import "net"

// ResolveHandler to handle the result of network address resolving.
type ResolveHandler interface {
	// HandleResolve handles the resolving result of host to addr.
	HandleResolve(host string, addr net.Addr)
}
