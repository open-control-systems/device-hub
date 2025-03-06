package devstore

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/system/sysmdns"
)

type testStoreMdnsHandlerDevice struct {
	typ  string
	desc string
}

type testStoreMdnsHandlerStore struct {
	err             error
	devices         map[string]testStoreMdnsHandlerDevice
	addCallCount    int
	removeCallCount int
}

func newTestStoreMdnsHandlerStore() *testStoreMdnsHandlerStore {
	return &testStoreMdnsHandlerStore{
		devices: make(map[string]testStoreMdnsHandlerDevice),
	}
}

func (s *testStoreMdnsHandlerStore) Add(uri string, typ string, desc string) error {
	if s.err != nil {
		return s.err
	}

	_, ok := s.devices[uri]
	if ok {
		return ErrDeviceExist
	}

	s.addCallCount++

	s.devices[uri] = testStoreMdnsHandlerDevice{
		typ:  typ,
		desc: desc,
	}

	return nil
}

func (s *testStoreMdnsHandlerStore) Remove(uri string) error {
	if s.err != nil {
		return s.err
	}

	s.removeCallCount++

	delete(s.devices, uri)

	return nil
}

func (*testStoreMdnsHandlerStore) GetDesc() []StoreItem {
	return []StoreItem{}
}

func (s *testStoreMdnsHandlerStore) count() int {
	return len(s.devices)
}

func (s *testStoreMdnsHandlerStore) checkDevice(uri string, typ string, desc string) bool {
	d, ok := s.devices[uri]
	if !ok {
		return false
	}

	return d.typ == typ && d.desc == desc
}

func TestStoreMdnsHandlerInvalidTxtRecordFormat(t *testing.T) {
	store := newTestStoreMdnsHandlerStore()
	mdnsHandler := NewStoreMdnsHandler(store)

	for _, record := range []string{
		"foo",
		"foo-bar",
		"",
		"foo=",
		"=foo",
		"=",
	} {
		service := &sysmdns.Service{
			TxtRecords: []string{record},
		}

		require.Nil(t, mdnsHandler.HandleService(service))
		require.Equal(t, 0, store.count())
	}
}

func TestStoreMdnsHandlerMissedRequiredTxtFields(t *testing.T) {
	store := newTestStoreMdnsHandlerStore()
	mdnsHandler := NewStoreMdnsHandler(store)

	for _, records := range [][]string{
		{
			"autodiscovery_mode=1",
		},
		{
			"autodiscovery_uri=http://bonsai-growlab.local/api/v1",
		},
		{
			"autodiscovery_desc=home-plant",
		},
		{
			"autodiscovery_mode=1",
			"autodiscovery_uri=http://bonsai-growlab.local/api/v1",
		},
		{
			"autodiscovery_uri=http://bonsai-growlab.local/api/v1",
			"autodiscovery_desc=home-plant",
		},
		{
			"autodiscovery_mode=1",
			"autodiscovery_desc=home-plant",
		},
	} {
		service := &sysmdns.Service{
			TxtRecords: records,
		}

		require.Nil(t, mdnsHandler.HandleService(service))
		require.Equal(t, 0, store.count())
	}
}

func TestStoreMdnsHandlerInvalidAutodiscoveryMode(t *testing.T) {
	store := newTestStoreMdnsHandlerStore()
	mdnsHandler := NewStoreMdnsHandler(store)

	for _, records := range [][]string{
		{
			"autodiscovery_mode=0",
			"autodiscovery_uri=http//bonsai-growlab.local/api/v1",
			"autodiscovery_desc=home-plant",
		},
		{
			"autodiscovery_mode=-1",
			"autodiscovery_uri=http//bonsai-growlab.local/api/v1",
			"autodiscovery_desc=home-plant",
		},
		{
			"autodiscovery_mode=2",
			"autodiscovery_uri=http//bonsai-growlab.local/api/v1",
			"autodiscovery_desc=home-plant",
		},
	} {
		service := &sysmdns.Service{
			TxtRecords: records,
		}

		require.Equal(t, status.StatusInvalidArg, mdnsHandler.HandleService(service))
		require.Equal(t, 0, store.count())
	}
}

func TestStoreMdnsHandlerFailedToAdd(t *testing.T) {
	store := newTestStoreMdnsHandlerStore()
	store.err = status.StatusTimeout

	mdnsHandler := NewStoreMdnsHandler(store)

	service := &sysmdns.Service{
		TxtRecords: []string{
			"autodiscovery_mode=1",
			"autodiscovery_uri=http//bonsai-growlab.local/api/v1",
			"autodiscovery_desc=home-plant",
			"autodiscovery_type=test-type",
		},
	}

	require.Equal(t, store.err, mdnsHandler.HandleService(service))
}

func TestStoreMdnsHandlerAddOK(t *testing.T) {
	store := newTestStoreMdnsHandlerStore()
	mdnsHandler := NewStoreMdnsHandler(store)

	service := &sysmdns.Service{
		TxtRecords: []string{
			"autodiscovery_mode=1",
			"autodiscovery_uri=http://bonsai-growlab.local/api/v1",
			"autodiscovery_desc=home-plant",
			"autodiscovery_type=test-type",
		},
	}

	require.Nil(t, mdnsHandler.HandleService(service))
	require.Equal(t, 1, store.count())

	require.True(t,
		store.checkDevice("http://bonsai-growlab.local/api/v1", "test-type", "home-plant"))
}

func TestStoreMdnsHandlerAddMultipleTimes(t *testing.T) {
	store := newTestStoreMdnsHandlerStore()
	mdnsHandler := NewStoreMdnsHandler(store)

	for n := 0; n < 10; n++ {
		service := &sysmdns.Service{
			TxtRecords: []string{
				"autodiscovery_mode=1",
				"autodiscovery_uri=http://bonsai-growlab.local/api/v1",
				"autodiscovery_desc=home-plant",
				"autodiscovery_type=test-type",
			},
		}

		require.Nil(t, mdnsHandler.HandleService(service))
	}

	require.Equal(t, 1, store.count())
	require.Equal(t, 1, store.addCallCount)

	require.True(t,
		store.checkDevice("http://bonsai-growlab.local/api/v1", "test-type", "home-plant"))
}
