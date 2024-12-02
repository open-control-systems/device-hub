package sysnet

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"golang.org/x/net/ipv4"

	"github.com/pion/mdns"
)

// It was decided to use the pure Go library for mDNS resolution, since the internal
// Go resolver behaves differently depending on the environment it's running in.
// For example, it can properly resolve mDNS addresses when running on the host machine,
// but fails to do so when running in the container. It's possible to force the Go runtime
// to work the same way by using CGO and os.Setenv("GODEBUG", "netdns=cgo"). CGO introduces
// other possible problems, and complicates the build process for different platforms.
type PionMdnsResolver struct {
	mu     sync.Mutex
	conn   *mdns.Conn
	closed bool
}

// Resolve mDNS hostname with pion library.
//
// Remarks:
//   - Can be used from multiple goroutines.
func (r *PionMdnsResolver) Resolve(ctx context.Context, hostname string) (net.Addr, error) {
	if !strings.Contains(hostname, ".local") {
		return nil, fmt.Errorf("pion-mdns-resolver: unsupported hostname: %s", hostname)
	}

	conn, err := r.getConn()
	if err != nil {
		return nil, err
	}

	_, addr, err := conn.Query(ctx, hostname)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

// Close the underlying mDNS connection.
func (r *PionMdnsResolver) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true

	if r.conn != nil {
		return r.conn.Close()
	}

	return nil
}

func (r *PionMdnsResolver) getConn() (*mdns.Conn, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("pion-mdns-resolver: closed")
	}

	if r.conn != nil {
		return r.conn, nil
	}

	// UDP Connection is closed when the mDNS connection is closed.
	udpConn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, fmt.Errorf("pion-mdns-resolver: failed to create UDP connection: %w", err)
	}

	packetConn := ipv4.NewPacketConn(udpConn)

	mdnsConn, err := mdns.Server(packetConn, &mdns.Config{})
	if err != nil {
		return nil, udpConn.Close()
	}

	r.conn = mdnsConn

	return mdnsConn, nil
}
