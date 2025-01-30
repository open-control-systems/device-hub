package sysnet

import (
	"context"
	"net"
	"sync"

	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// ResolveStore caches the result of hostname resolving.
type ResolveStore struct {
	updateCh chan struct{}

	mu            sync.Mutex
	knownHosts    map[string]struct{}
	resolvedAddrs map[string]net.Addr
}

// NewResolveStore is an initialization of ResolveStore.
func NewResolveStore() *ResolveStore {
	return &ResolveStore{
		updateCh:      make(chan struct{}, 1),
		knownHosts:    make(map[string]struct{}),
		resolvedAddrs: make(map[string]net.Addr),
	}
}

// HandleResolve caches known resolved addresses.
//
// Remarks:
//   - Unknown hosts are filtered out.
func (s *ResolveStore) HandleResolve(hostname string, addr net.Addr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.knownHosts[hostname]; !ok {
		return
	}

	ra, ok := s.resolvedAddrs[hostname]
	if !ok {
		syscore.LogInf.Printf("resolve-store: addr resolved: hostname=%s: addr=%s\n",
			hostname, addr)

		s.resolvedAddrs[hostname] = addr
	} else if ra.String() != addr.String() {
		syscore.LogInf.Printf("resolve-store: addr changed for hostname=%s: cur=%s new=%s\n",
			hostname, ra, addr)

		s.resolvedAddrs[hostname] = addr
	}

	select {
	case s.updateCh <- struct{}{}:
	default:
	}
}

// Resolve resolves the hostname to the network address.
//
// Remarks:
//   - Resolving an unknown hostname will always fail.
func (s *ResolveStore) Resolve(ctx context.Context, hostname string) (net.Addr, error) {
	if addr, err := s.getAddr(hostname); err == nil {
		return addr, nil
	}

	return s.waitAddr(ctx, hostname)
}

// Add adds hostname to the list of known hosts.
func (s *ResolveStore) Add(hostname string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.knownHosts[hostname] = struct{}{}
}

// Remove removes hostname from the list of known hosts.
func (s *ResolveStore) Remove(hostname string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.knownHosts, hostname)
	delete(s.resolvedAddrs, hostname)
}

func (s *ResolveStore) getAddr(hostname string) (net.Addr, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr, ok := s.resolvedAddrs[hostname]
	if !ok {
		return nil, status.StatusNoData
	}

	return addr, nil
}

func (s *ResolveStore) waitAddr(ctx context.Context, hostname string) (net.Addr, error) {
	for {
		select {
		case <-s.updateCh:
			return s.getAddr(hostname)

		case <-ctx.Done():
			return nil, status.StatusTimeout
		}
	}
}
