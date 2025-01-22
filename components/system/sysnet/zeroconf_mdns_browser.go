package sysnet

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"

	"github.com/open-control-systems/device-hub/components/core"
)

// ZeroconfMdnsBrowserParams represents various options for zeroconf mDNS browser.
type ZeroconfMdnsBrowserParams struct {
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

// ZeroconfMdnsBrowser browses the local network for the mDNS devices.
//
// References:
//   - https://github.com/grandcat/zeroconf
type ZeroconfMdnsBrowser struct {
	params   ZeroconfMdnsBrowserParams
	ctx      context.Context
	handler  ResolveHandler
	resolver *zeroconf.Resolver
}

// NewZeroconfMdnsBrowser is an initialization of ZeroconfMdnsBrowser.
func NewZeroconfMdnsBrowser(
	ctx context.Context,
	handler ResolveHandler,
	params ZeroconfMdnsBrowserParams,
) (*ZeroconfMdnsBrowser, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	return &ZeroconfMdnsBrowser{
		params:   params,
		ctx:      ctx,
		handler:  handler,
		resolver: resolver,
	}, nil
}

// Run executes a single mDNS lookup operation.
func (b *ZeroconfMdnsBrowser) Run() error {
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
func (*ZeroconfMdnsBrowser) Close() error {
	return nil
}

// ReportError reports browsing errors to the log.
func (b *ZeroconfMdnsBrowser) ReportError(err error) {
	core.LogErr.Printf("mdns-zeroconf-browser: browsing failed: service=%s domain=%s: %v\n",
		b.params.Service, b.params.Domain, err)
}

func (b *ZeroconfMdnsBrowser) handleEntry(entry *zeroconf.ServiceEntry) {
	if len(entry.AddrIPv4) < 1 {
		core.LogWrn.Printf("mdns-zeroconf-browser: ignore entry: service=%s domain=%s:"+
			" IPv4 address not found\n", b.params.Service, b.params.Domain)
	}

	b.handler.HandleResolve(
		strings.TrimSuffix(entry.HostName, "."),
		&net.IPAddr{IP: entry.AddrIPv4[0]},
	)
}
