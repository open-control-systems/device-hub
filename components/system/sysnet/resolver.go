package sysnet

import (
	"context"
	"net"
)

// Resolver to resolve network addresses.
type Resolver interface {
	// Resolve hostname.
	Resolve(ctx context.Context, hostname string) (net.Addr, error)
}
