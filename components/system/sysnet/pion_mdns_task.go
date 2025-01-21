package sysnet

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/ipv4"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/pion/mdns"
)

// PionMdnsTask implements mDNS resolving.
//
// It was decided to use the pure Go library for mDNS resolution, since the internal
// Go resolver behaves differently depending on the environment it's running in.
// For example, it can properly resolve mDNS addresses when running on the host machine,
// but fails to do so when running in the container. It's possible to force the Go runtime
// to work the same way by using CGO and os.Setenv("GODEBUG", "netdns=cgo"). CGO introduces
// other possible problems, and complicates the build process for different platforms.
type PionMdnsTask struct {
	host    string
	timeout time.Duration
	ctx     context.Context
	handler ResolveHandler
	mu      sync.Mutex
	conn    *mdns.Conn
	closed  bool
}

// NewPionMdnsTask is an initialization of PionMdnsTask.
func NewPionMdnsTask(
	ctx context.Context,
	handler ResolveHandler,
	timeout time.Duration,
	host string,
) *PionMdnsTask {
	if !strings.Contains(host, ".local") {
		panic(fmt.Sprintf("pion-mdns-task: unsupported host: %s", host))
	}

	return &PionMdnsTask{
		host:    host,
		timeout: timeout,
		ctx:     ctx,
		handler: handler,
	}
}

// ReportError reports errors from the Run() call.
func (t *PionMdnsTask) ReportError(err error) {
	core.LogErr.Printf("pion-mdns-task: operation failed: host=%s err=%v\n", t.host, err)
}

// Run performs mDNS resolving for the configured host.
func (t *PionMdnsTask) Run() error {
	conn, err := t.getConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(t.ctx, t.timeout)
	defer cancel()

	_, addr, err := conn.Query(ctx, t.host)
	if err != nil {
		return err
	}

	t.handler.HandleResolve(t.host, addr)

	return nil
}

// Close closes the underlying mDNS connection.
func (t *PionMdnsTask) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.closed = true

	if t.conn != nil {
		return t.conn.Close()
	}

	return nil
}

func (t *PionMdnsTask) getConn() (*mdns.Conn, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, fmt.Errorf("pion-mdns-task: closed")
	}

	if t.conn != nil {
		return t.conn, nil
	}

	// UDP Connection is closed when the mDNS connection is closed.
	udpConn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, fmt.Errorf("pion-mdns-task: failed to create UDP connection: %w", err)
	}

	packetConn := ipv4.NewPacketConn(udpConn)

	mdnsConn, err := mdns.Server(packetConn, &mdns.Config{})
	if err != nil {
		return nil, udpConn.Close()
	}

	t.conn = mdnsConn

	return mdnsConn, nil
}
