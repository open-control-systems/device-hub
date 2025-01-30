package sysmdns

import (
	"context"
	"time"

	"github.com/open-control-systems/zeroconf"

	"github.com/open-control-systems/device-hub/components/system/syscore"
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

	// Opts is a zeroconf browse configuration options.
	Opts []zeroconf.ClientOption
}

// ZeroconfBrowser browses the local network for the mDNS devices.
//
// References:
//   - https://github.com/grandcat/zeroconf
type ZeroconfBrowser struct {
	params  ZeroconfBrowserParams
	ctx     context.Context
	handler ServiceHandler
}

// NewZeroconfBrowser is an initialization of ZeroconfBrowser.
func NewZeroconfBrowser(
	ctx context.Context,
	handler ServiceHandler,
	params ZeroconfBrowserParams,
) *ZeroconfBrowser {
	return &ZeroconfBrowser{
		params:  params,
		ctx:     ctx,
		handler: handler,
	}
}

// Run executes a single mDNS lookup operation.
func (b *ZeroconfBrowser) Run() error {
	resolver, err := zeroconf.NewResolver(b.params.Opts...)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(b.ctx, b.params.Timeout)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry)

	if err := resolver.Browse(ctx, b.params.Service, b.params.Domain, entries); err != nil {
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

// Stop closes the browser resources.
func (*ZeroconfBrowser) Stop() error {
	return nil
}

// HandleError handles browsing errors.
func (b *ZeroconfBrowser) HandleError(err error) {
	syscore.LogErr.Printf("mdns-zeroconf-browser: browsing failed: service=%s domain=%s: %v\n",
		b.params.Service, b.params.Domain, err)
}

func (b *ZeroconfBrowser) handleEntry(entry *zeroconf.ServiceEntry) {
	service := &Service{
		Instance:   entry.Instance,
		Name:       entry.Service,
		Hostname:   entry.HostName,
		Port:       entry.Port,
		TxtRecords: entry.Text,
		AddrsIPv4:  entry.AddrIPv4,
		AddrsIPv6:  entry.AddrIPv6,
	}

	if err := b.handler.HandleService(service); err != nil {
		syscore.LogWrn.Printf("mdns-zeroconf-browser: failed to handle service: service=%s"+
			" domain=%s err=%v\n", b.params.Service, b.params.Domain, err)
	}
}
