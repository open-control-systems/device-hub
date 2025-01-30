package sysnet

import (
	"context"
	"net"
)

// Resolver to resolve a resource hostname.
type Resolver interface {
	// Resolve resolves a resource hostname to the network address.
	//
	// Examples:
	//   - google.com -> 142.251.208.110
	//   - bonsai-growlab.local -> 192.168.1.4
	Resolve(ctx context.Context, hostname string) (net.Addr, error)
}
