package devstore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testStoreAliveMonitorClock struct {
	now time.Time
}

func (c *testStoreAliveMonitorClock) Now() time.Time {
	return c.now
}

type testStoreAliveMonitorDevice struct {
	typ  string
	desc string
}

type testStoreAliveMonitorStore struct {
	err             error
	devices         map[string]testStoreAliveMonitorDevice
	addCallCount    int
	removeCallCount int
}

func newTestStoreAliveMonitorStore() *testStoreAliveMonitorStore {
	return &testStoreAliveMonitorStore{
		devices: make(map[string]testStoreAliveMonitorDevice),
	}
}

func (s *testStoreAliveMonitorStore) Add(uri string, typ string, desc string) error {
	s.addCallCount++

	if s.err != nil {
		return s.err
	}

	s.devices[uri] = testStoreAliveMonitorDevice{
		typ:  typ,
		desc: desc,
	}

	return nil
}

func (s *testStoreAliveMonitorStore) Remove(uri string) error {
	s.removeCallCount++

	if s.err != nil {
		return s.err
	}

	delete(s.devices, uri)

	return nil
}

func (s *testStoreAliveMonitorStore) GetDesc() []StoreItem {
	var ret []StoreItem

	for uri, device := range s.devices {
		ret = append(ret, StoreItem{
			URI:  uri,
			Type: device.typ,
			Desc: device.desc,
		})
	}

	return ret
}

func (s *testStoreAliveMonitorStore) count() int {
	return len(s.devices)
}

func (s *testStoreAliveMonitorStore) checkDevice(uri string, typ string, desc string) bool {
	d, ok := s.devices[uri]
	if !ok {
		return false
	}

	return d.typ == typ && d.desc == desc
}

func TestStoreAliveMonitorVerifyInactivityOK(t *testing.T) {
	inactiveInterval := time.Minute

	uri := "http://bonsai-growlab.local/api/v1"
	desc := "home-plant"
	typ := "test-type"

	clock := &testStoreAliveMonitorClock{}
	store := newTestStoreAliveMonitorStore()

	monitor := NewStoreAliveMonitor(clock, store, inactiveInterval)

	require.Nil(t, monitor.Add(uri, typ, desc))
	require.Nil(t, monitor.Run())

	require.Equal(t, 1, store.count())
	require.Equal(t, 1, store.addCallCount)
	require.Equal(t, 0, store.removeCallCount)
	require.True(t, store.checkDevice(uri, typ, desc))

	clock.now = clock.now.Add(inactiveInterval / 2)
	require.Nil(t, monitor.Run())

	require.Equal(t, 1, store.count())
	require.Equal(t, 1, store.addCallCount)
	require.Equal(t, 0, store.removeCallCount)
	require.True(t, store.checkDevice(uri, typ, desc))
}

func TestStoreAliveMonitorVerifyInactivityTimeout(t *testing.T) {
	resolution := time.Minute
	inactiveInterval := resolution * 10

	uri := "http://bonsai-growlab.local/api/v1"
	desc := "home-plant"
	typ := "test-type"

	clock := &testStoreAliveMonitorClock{}
	store := newTestStoreAliveMonitorStore()

	monitor := NewStoreAliveMonitor(clock, store, inactiveInterval)

	notifier := monitor.Monitor(uri)
	require.NotNil(t, notifier)

	require.Nil(t, monitor.Add(uri, typ, desc))
	require.Nil(t, monitor.Run())

	require.Equal(t, 1, store.count())
	require.Equal(t, 1, store.addCallCount)
	require.Equal(t, 0, store.removeCallCount)
	require.True(t, store.checkDevice(uri, typ, desc))

	clock.now = clock.now.Add(inactiveInterval - resolution)
	require.Nil(t, monitor.Run())

	notifier.NotifyAlive()

	require.Equal(t, 1, store.addCallCount)
	require.Equal(t, 0, store.removeCallCount)
	require.True(t, store.checkDevice(uri, typ, desc))

	clock.now = clock.now.Add(resolution)
	require.Nil(t, monitor.Run())

	require.Equal(t, 1, store.addCallCount)
	require.Equal(t, 0, store.removeCallCount)
	require.True(t, store.checkDevice(uri, typ, desc))

	clock.now = clock.now.Add(inactiveInterval - resolution)
	require.Nil(t, monitor.Run())

	require.Equal(t, 1, store.addCallCount)
	require.Equal(t, 1, store.removeCallCount)
	require.False(t, store.checkDevice(uri, typ, desc))

	require.Nil(t, monitor.Run())

	require.Equal(t, 1, store.addCallCount)
	require.Equal(t, 1, store.removeCallCount)
	require.False(t, store.checkDevice(uri, typ, desc))
}

func TestStoreAliveMonitorVerifyInactivityOnRestore(t *testing.T) {
	resolution := time.Minute
	inactiveInterval := resolution * 10

	uri := "http://bonsai-growlab.local/api/v1"
	desc := "home-plant"
	typ := "test-type"

	clock := &testStoreAliveMonitorClock{}

	store := newTestStoreAliveMonitorStore()
	require.Nil(t, store.Add(uri, typ, desc))

	monitor := NewStoreAliveMonitor(clock, store, inactiveInterval)

	clock.now = clock.now.Add(inactiveInterval)

	require.Nil(t, monitor.Run())
	require.Equal(t, 1, store.addCallCount)
	require.Equal(t, 1, store.removeCallCount)
	require.False(t, store.checkDevice(uri, typ, desc))

	require.Nil(t, monitor.Run())
	require.Equal(t, 1, store.addCallCount)
	require.Equal(t, 1, store.removeCallCount)
	require.False(t, store.checkDevice(uri, typ, desc))
}
