package sysnet

import (
	"context"
	"net"
)

// Resolver to resolve network addresses.
type Resolver interface {
	// Resolve resolves network address.
	Resolve(ctx context.Context, address string) (net.Addr, error)
}
