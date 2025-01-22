package sysmdns

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
)

// ZeroconfBrowserParams represents various options for zeroconf mDNS browser.
type ZeroconfBrowserParams struct {
	// Service is a mDNS service to lookup for.
	//
	// Examples:
	//  - Lookup for all HTTP services over TCP protocol: "_http._tcp".
	Service string

	// Domain is a mDNS domain.
	//
	// Examples:
	//  - Local domain: "local".
	Domain string

	// Timeout is a mDNS browsing timeout.
	Timeout time.Duration
}

// ZeroconfBrowser browses the local network for the mDNS devices.
//
// References:
//   - https://github.com/grandcat/zeroconf
type ZeroconfBrowser struct {
	params   ZeroconfBrowserParams
	ctx      context.Context
	handler  sysnet.ResolveHandler
	resolver *zeroconf.Resolver
}

// NewZeroconfBrowser is an initialization of ZeroconfBrowser.
func NewZeroconfBrowser(
	ctx context.Context,
	handler sysnet.ResolveHandler,
	params ZeroconfBrowserParams,
) (*ZeroconfBrowser, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	return &ZeroconfBrowser{
		params:   params,
		ctx:      ctx,
		handler:  handler,
		resolver: resolver,
	}, nil
}

// Run executes a single mDNS lookup operation.
func (b *ZeroconfBrowser) Run() error {
	ctx, cancel := context.WithTimeout(b.ctx, b.params.Timeout)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry)

	if err := b.resolver.Browse(ctx, b.params.Service, b.params.Domain, entries); err != nil {
		return err
	}

	for {
		select {
		case entry := <-entries:
			b.handleEntry(entry)

		case <-ctx.Done():
			return nil
		}
	}
}

// Close closes the browser resources.
func (*ZeroconfBrowser) Close() error {
	return nil
}

// HandleError handles browsing errors.
func (b *ZeroconfBrowser) HandleError(err error) {
	core.LogErr.Printf("mdns-zeroconf-browser: browsing failed: service=%s domain=%s: %v\n",
		b.params.Service, b.params.Domain, err)
}

func (b *ZeroconfBrowser) handleEntry(entry *zeroconf.ServiceEntry) {
	if len(entry.AddrIPv4) < 1 {
		core.LogWrn.Printf("mdns-zeroconf-browser: ignore entry: service=%s domain=%s:"+
			" IPv4 address not found\n", b.params.Service, b.params.Domain)
	}

	b.handler.HandleResolve(
		strings.TrimSuffix(entry.HostName, "."),
		&net.IPAddr{IP: entry.AddrIPv4[0]},
	)
}
