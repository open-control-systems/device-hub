package pipdevice

import (
	"net"
	"testing"

	"github.com/open-control-systems/device-hub/components/status"
	"github.com/stretchr/testify/require"
)

type testStoreMdnsHandlerStore struct {
	err             error
	devices         map[string]string
	addCallCount    int
	removeCallCount int
}

func newTestStoreMdnsHandlerStore() *testStoreMdnsHandlerStore {
	return &testStoreMdnsHandlerStore{
		devices: make(map[string]string),
	}
}

func (s *testStoreMdnsHandlerStore) Add(uri string, desc string) error {
	if s.err != nil {
		return s.err
	}

	_, ok := s.devices[uri]
	if ok {
		return ErrDeviceExist
	}

	s.addCallCount++
	s.devices[uri] = desc

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

func (s *testStoreMdnsHandlerStore) checkDevice(uri string, desc string) bool {
	d, ok := s.devices[uri]
	if !ok {
		return false
	}

	return d == desc
}

type testStoreMdnsHandlerMdnsService struct {
	instance   string
	name       string
	hostname   string
	port       int
	txtRecords []string
	addrs      []net.IP
}

func (s *testStoreMdnsHandlerMdnsService) Instance() string {
	return s.instance
}

func (s *testStoreMdnsHandlerMdnsService) Name() string {
	return s.name
}

func (s *testStoreMdnsHandlerMdnsService) Hostname() string {
	return s.hostname
}

func (s *testStoreMdnsHandlerMdnsService) Port() int {
	return s.port
}

func (s *testStoreMdnsHandlerMdnsService) TxtRecords() []string {
	return s.txtRecords
}

func (s *testStoreMdnsHandlerMdnsService) Addrs() []net.IP {
	return s.addrs
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
		service := &testStoreMdnsHandlerMdnsService{}
		service.txtRecords = append(service.txtRecords, record)

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
		service := &testStoreMdnsHandlerMdnsService{}
		service.txtRecords = append(service.txtRecords, records...)

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
		service := &testStoreMdnsHandlerMdnsService{}
		service.txtRecords = append(service.txtRecords, records...)

		require.Equal(t, status.StatusInvalidArg, mdnsHandler.HandleService(service))
		require.Equal(t, 0, store.count())
	}
}

func TestStoreMdnsHandlerFailedToAdd(t *testing.T) {
	store := newTestStoreMdnsHandlerStore()
	store.err = status.StatusTimeout

	mdnsHandler := NewStoreMdnsHandler(store)

	service := &testStoreMdnsHandlerMdnsService{}
	service.txtRecords = []string{
		"autodiscovery_mode=1",
		"autodiscovery_uri=http//bonsai-growlab.local/api/v1",
		"autodiscovery_desc=home-plant",
	}

	require.Equal(t, store.err, mdnsHandler.HandleService(service))
}

func TestStoreMdnsHandlerAddOK(t *testing.T) {
	store := newTestStoreMdnsHandlerStore()
	mdnsHandler := NewStoreMdnsHandler(store)

	service := &testStoreMdnsHandlerMdnsService{}
	service.txtRecords = []string{
		"autodiscovery_mode=1",
		"autodiscovery_uri=http://bonsai-growlab.local/api/v1",
		"autodiscovery_desc=home-plant",
	}

	require.Nil(t, mdnsHandler.HandleService(service))
	require.Equal(t, 1, store.count())
	require.True(t, store.checkDevice("http://bonsai-growlab.local/api/v1", "home-plant"))
}

func TestStoreMdnsHandlerAddMultipleTimes(t *testing.T) {
	store := newTestStoreMdnsHandlerStore()
	mdnsHandler := NewStoreMdnsHandler(store)

	for n := 0; n < 10; n++ {
		service := &testStoreMdnsHandlerMdnsService{}
		service.txtRecords = []string{
			"autodiscovery_mode=1",
			"autodiscovery_uri=http://bonsai-growlab.local/api/v1",
			"autodiscovery_desc=home-plant",
		}

		require.Nil(t, mdnsHandler.HandleService(service))
	}

	require.Equal(t, 1, store.count())
	require.Equal(t, 1, store.addCallCount)
	require.True(t, store.checkDevice("http://bonsai-growlab.local/api/v1", "home-plant"))
}
