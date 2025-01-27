package sysmdns

import (
	"context"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
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
	resolver, err := zeroconf.NewResolver(nil)
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
	service := &zeroconfService{entry: entry}

	if err := b.handler.HandleService(service); err != nil {
		syscore.LogWrn.Printf("mdns-zeroconf-browser: failed to handle service: service=%s"+
			" domain=%s err=%v\n", b.params.Service, b.params.Domain, err)
	}
}

type zeroconfService struct {
	entry *zeroconf.ServiceEntry
}

func (s *zeroconfService) Instance() string {
	return s.entry.Instance
}

func (s *zeroconfService) Name() string {
	return s.entry.Service
}

func (s *zeroconfService) Hostname() string {
	return s.entry.HostName
}

func (s *zeroconfService) Port() int {
	return s.entry.Port
}

func (s *zeroconfService) TxtRecords() []string {
	return s.entry.Text
}

func (s *zeroconfService) Addrs() []net.IP {
	return s.entry.AddrIPv4
}
