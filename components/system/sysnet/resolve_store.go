package sysnet

import (
	"context"
	"net"
	"sync"

	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// ResolveStore caches the result of host resolving.
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
func (s *ResolveStore) HandleResolve(host string, addr net.Addr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.knownHosts[host]; !ok {
		return
	}

	ra, ok := s.resolvedAddrs[host]
	if !ok {
		syscore.LogInf.Printf("resolve-store: addr resolved: host=%s: addr=%s\n",
			host, addr)

		s.resolvedAddrs[host] = addr
	} else if ra.String() != addr.String() {
		syscore.LogInf.Printf("resolve-store: addr changed for host=%s: cur=%s new=%s\n",
			host, ra, addr)

		s.resolvedAddrs[host] = addr
	}

	select {
	case s.updateCh <- struct{}{}:
	default:
	}
}

// Resolve resolves the host address to the network address.
//
// Remarks:
//   - Resolving an unknown host will always fail.
func (s *ResolveStore) Resolve(ctx context.Context, host string) (net.Addr, error) {
	if addr, err := s.getAddr(host); err == nil {
		return addr, nil
	}

	return s.waitAddr(ctx, host)
}

// Add adds host to the list of known hosts.
func (s *ResolveStore) Add(host string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.knownHosts[host] = struct{}{}
}

// Remove removes host from the list of known hosts.
func (s *ResolveStore) Remove(host string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.knownHosts, host)
	delete(s.resolvedAddrs, host)
}

func (s *ResolveStore) getAddr(host string) (net.Addr, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr, ok := s.resolvedAddrs[host]
	if !ok {
		return nil, status.StatusNoData
	}

	return addr, nil
}

func (s *ResolveStore) waitAddr(ctx context.Context, host string) (net.Addr, error) {
	for {
		select {
		case <-s.updateCh:
			return s.getAddr(host)

		case <-ctx.Done():
			return nil, status.StatusTimeout
		}
	}
}
