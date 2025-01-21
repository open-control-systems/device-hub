package sysnet

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-control-systems/device-hub/components/status"
)

func TestResolveStoreResolveContexTimeout(t *testing.T) {
	store := NewResolveStore()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	addr, err := store.Resolve(ctx, "foo.bar.local")
	require.Nil(t, addr)
	require.Equal(t, status.StatusTimeout, err)
}

func TestResolveStoreResolveHandleResolveFiltered(t *testing.T) {
	store := NewResolveStore()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	mdnsAddr := "foo.bar.local"
	netAddr := net.IPAddr{IP: net.IPv4(192, 168, 4, 2)}

	addr, err := store.Resolve(ctx, mdnsAddr)
	require.Nil(t, addr)
	require.Equal(t, status.StatusTimeout, err)

	store.HandleResolve(mdnsAddr, &netAddr)

	addr, err = store.Resolve(ctx, mdnsAddr)
	require.Nil(t, addr)
	require.Equal(t, status.StatusTimeout, err)
}

func TestResolveStoreResolveHandleResolve(t *testing.T) {
	store := NewResolveStore()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	mdnsAddr := "foo.bar.local"
	netAddr := net.IPAddr{IP: net.IPv4(192, 168, 4, 2)}

	store.Add(mdnsAddr)
	store.HandleResolve(mdnsAddr, &netAddr)

	addr, err := store.Resolve(ctx, mdnsAddr)
	require.Nil(t, err)
	require.Equal(t, netAddr.String(), addr.String())
}

func TestResolveStoreResolveHandleResolveAsync(t *testing.T) {
	store := NewResolveStore()

	mdnsAddr := "foo.bar.local"
	netAddr := net.IPAddr{IP: net.IPv4(192, 168, 4, 2)}

	store.Add(mdnsAddr)

	go func() {
		time.Sleep(time.Millisecond * 300)
		store.HandleResolve(mdnsAddr, &netAddr)
	}()

	addr, err := store.Resolve(context.Background(), mdnsAddr)
	require.Nil(t, err)
	require.Equal(t, netAddr.String(), addr.String())
}

func TestResolveStoreResolveHandleResolveAddrChanged(t *testing.T) {
	store := NewResolveStore()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	mdnsAddr := "foo.bar.local"

	curNetAddr := net.IPAddr{IP: net.IPv4(192, 168, 4, 2)}
	newNetAddr := net.IPAddr{IP: net.IPv4(192, 168, 4, 1)}
	require.NotEqual(t, curNetAddr.String(), newNetAddr.String())

	store.Add(mdnsAddr)

	store.HandleResolve(mdnsAddr, &curNetAddr)
	addr, err := store.Resolve(ctx, mdnsAddr)
	require.Nil(t, err)
	require.Equal(t, curNetAddr.String(), addr.String())

	store.HandleResolve(mdnsAddr, &newNetAddr)
	addr, err = store.Resolve(ctx, mdnsAddr)
	require.Nil(t, err)
	require.Equal(t, newNetAddr.String(), addr.String())
}

func TestResolveStoreResolveAfterRemove(t *testing.T) {
	store := NewResolveStore()

	mdnsAddr := "foo.bar.local"
	netAddr := net.IPAddr{IP: net.IPv4(192, 168, 4, 2)}

	store.Add(mdnsAddr)
	store.HandleResolve(mdnsAddr, &netAddr)

	addr, err := store.Resolve(context.Background(), mdnsAddr)
	require.Nil(t, err)
	require.Equal(t, netAddr.String(), addr.String())

	store.Remove(mdnsAddr)
	addr, err = store.Resolve(context.Background(), mdnsAddr)
	require.Equal(t, status.StatusNoData, err)
	require.Nil(t, addr)
}
