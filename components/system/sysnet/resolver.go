package sysnet

import (
	"context"
	"net"
)

type Resolver interface {
	// Resolve hostname.
	Resolve(ctx context.Context, hostname string) (net.Addr, error)
}
