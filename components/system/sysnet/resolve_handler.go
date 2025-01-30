package sysnet

import "net"

// ResolveHandler to handle the result of network address resolving.
type ResolveHandler interface {
	// HandleResolve handles the resolving result of hostname to addr.
	HandleResolve(hostname string, addr net.Addr)
}
