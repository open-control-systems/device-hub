package devstore

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-control-systems/device-hub/components/device/devcore"
	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/storage/stcore"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
)

type testCacheStoreDB struct {
	data map[string][]byte
}

func newTestCacheStoreDB() *testCacheStoreDB {
	return &testCacheStoreDB{
		data: make(map[string][]byte),
	}
}

func (d *testCacheStoreDB) Read(key string) ([]byte, error) {
	buf, ok := d.data[key]
	if !ok {
		return []byte{}, status.StatusNoData
	}

	return buf, nil
}

func (d *testCacheStoreDB) Write(key string, buf []byte) error {
	b := make([]byte, len(buf))
	copy(b, buf)

	d.data[key] = b

	return nil
}

func (d *testCacheStoreDB) Remove(key string) error {
	delete(d.data, key)

	return nil
}

func (d *testCacheStoreDB) ForEach(fn func(key string, b []byte) error) error {
	for k, v := range d.data {
		if err := fn(k, v); err != nil {
			return err
		}
	}

	return nil
}

func (*testCacheStoreDB) Close() error {
	return nil
}

func (d *testCacheStoreDB) count() int {
	return len(d.data)
}

type testCacheStoreDataHandler struct {
	telemetry    chan devcore.JSON
	registration chan devcore.JSON
}

func newTestCacheStoreDataHandler() *testCacheStoreDataHandler {
	return &testCacheStoreDataHandler{
		telemetry:    make(chan devcore.JSON),
		registration: make(chan devcore.JSON),
	}
}

func (h *testCacheStoreDataHandler) HandleTelemetry(_ string, js devcore.JSON) error {
	select {
	case h.telemetry <- maps.Clone(js):
	default:
	}

	return nil
}

func (h *testCacheStoreDataHandler) HandleRegistration(_ string, js devcore.JSON) error {
	select {
	case h.registration <- maps.Clone(js):
	default:
	}

	return nil
}

type testCacheStoreClock struct {
	timestamp int64
}

func (c *testCacheStoreClock) SetTimestamp(timestamp int64) error {
	c.timestamp = timestamp

	return nil
}

func (c *testCacheStoreClock) GetTimestamp() (int64, error) {
	return c.timestamp, nil
}

type testCacheStoreHTTPDataHandler struct {
	js devcore.JSON
}

func newTestCacheStoreHTTPDataHandler(data devcore.JSON) *testCacheStoreHTTPDataHandler {
	return &testCacheStoreHTTPDataHandler{
		js: maps.Clone(data),
	}
}

func (h *testCacheStoreHTTPDataHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(h.js); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TestCacheStoreStartStopEmpty(t *testing.T) {
	db := newTestCacheStoreDB()
	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	require.Nil(t, store.Start())
}

func TestCacheStoreStopNoStart(t *testing.T) {
	db := newTestCacheStoreDB()
	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()
}

func TestCacheStoreGetDescEmpty(t *testing.T) {
	db := newTestCacheStoreDB()
	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	descs := store.GetDesc()
	require.Empty(t, descs)
}

func TestCacheStoreRemoveNoAdd(t *testing.T) {
	db := newTestCacheStoreDB()
	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	require.Equal(t, status.StatusNoData, store.Remove("foo-bar-baz"))
}

func TestCacheStoreAddURIUnsupportedScheme(t *testing.T) {
	db := newTestCacheStoreDB()
	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	require.Equal(t, status.StatusNotSupported,
		store.Add("foo-bar-baz", "test-type", "foo-bar-baz"))
}

func TestCacheStoreAddRemoveResourceNoResponse(t *testing.T) {
	db := newTestCacheStoreDB()
	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	ctx, cancelFunc := context.WithTimeoutCause(
		context.Background(),
		time.Millisecond*500,
		status.StatusTimeout,
	)
	defer cancelFunc()

	store := NewCacheStore(
		ctx,
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	tests := []struct {
		uri  string
		typ  string
		desc string
	}{
		{"http://devcore.example.com:123/api/v10", "test-type", "foo-bar-baz"},
		{"http://192.1.2.3:8787/api/v3", "test-type", "foo-bar-baz"},
		{"https://192.1.2.3:1234", "test-type", "foo-bar-baz"},
		{"http://bonsai-growlab.local:234/api/v1", "test-type", "foo-bar-baz"},
	}

	for _, test := range tests {
		require.Nil(t, store.Add(test.uri, test.typ, test.desc))
	}

	<-ctx.Done()
	require.Equal(t, status.StatusTimeout, context.Cause(ctx))

	for _, test := range tests {
		found := false

		for _, desc := range store.GetDesc() {
			if desc.URI == test.uri && desc.Desc == test.desc {
				found = true
			}
		}

		require.True(t, found)
	}

	require.Equal(t, len(tests), db.count())

	for _, test := range tests {
		require.Nil(t, store.Remove(test.uri))
	}

	require.Equal(t, 0, db.count())
}

func TestCacheStoreAddRemove(t *testing.T) {
	db := newTestCacheStoreDB()
	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	deviceID := "0xABCD"

	telemetryData := make(devcore.JSON)
	telemetryData["timestamp"] = float64(123)
	telemetryData["temperature"] = float64(123.222)

	registrationData := make(devcore.JSON)
	registrationData["timestamp"] = float64(123)
	registrationData["device_id"] = deviceID

	telemetryHandler := newTestCacheStoreHTTPDataHandler(telemetryData)
	registrationHandler := newTestCacheStoreHTTPDataHandler(registrationData)

	mux := http.NewServeMux()
	mux.Handle("/telemetry", telemetryHandler)
	mux.Handle("/registration", registrationHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	require.Nil(t, store.Add(server.URL, "test-type", "foo-bar-baz"))

	require.True(t, maps.Equal(telemetryData, <-handler.telemetry))
	require.True(t, maps.Equal(registrationData, <-handler.registration))
}

func TestCacheStoreRestore(t *testing.T) {
	db := newTestCacheStoreDB()

	makeStore := func(d stcore.DB, h devcore.DataHandler) *CacheStore {
		clock := &testCacheStoreClock{}

		storeParams := CacheStoreParams{}
		storeParams.HTTP.FetchInterval = time.Millisecond * 100
		storeParams.HTTP.FetchTimeout = time.Millisecond * 100

		return NewCacheStore(
			context.Background(),
			clock,
			clock,
			h,
			d,
			sysnet.NewResolveStore(),
			storeParams,
		)
	}

	handler1 := newTestCacheStoreDataHandler()
	store1 := makeStore(db, handler1)

	require.Empty(t, store1.GetDesc())

	deviceID := "0xABCD"

	telemetryData := make(devcore.JSON)
	telemetryData["timestamp"] = float64(123)
	telemetryData["temperature"] = float64(123.222)

	registrationData := make(devcore.JSON)
	registrationData["timestamp"] = float64(123)
	registrationData["device_id"] = deviceID

	telemetryHandler := newTestCacheStoreHTTPDataHandler(telemetryData)
	registrationHandler := newTestCacheStoreHTTPDataHandler(registrationData)

	mux := http.NewServeMux()
	mux.Handle("/telemetry", telemetryHandler)
	mux.Handle("/registration", registrationHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	deviceURI := server.URL
	deviceDesc := "foo-bar-baz"
	deviceType := "test-type"

	require.Nil(t, store1.Add(deviceURI, deviceType, deviceDesc))

	require.True(t, maps.Equal(telemetryData, <-handler1.telemetry))
	require.True(t, maps.Equal(registrationData, <-handler1.registration))

	require.Nil(t, store1.Stop())

	handler2 := newTestCacheStoreDataHandler()
	store2 := makeStore(db, handler2)

	descs := store2.GetDesc()
	require.Equal(t, 1, len(descs))

	desc := descs[0]
	require.Equal(t, deviceURI, desc.URI)
	require.Equal(t, deviceDesc, desc.Desc)
	require.Equal(t, deviceType, desc.Type)

	require.Nil(t, store2.Start())

	require.NotNil(t, store2.Add(deviceURI, deviceType, deviceDesc))
	require.True(t, maps.Equal(telemetryData, <-handler2.telemetry))
	require.True(t, maps.Equal(registrationData, <-handler2.registration))

	require.Nil(t, store2.Remove(deviceURI))

	handler3 := newTestCacheStoreDataHandler()
	store3 := makeStore(db, handler3)

	require.Nil(t, store3.Add(deviceURI, deviceType, deviceDesc))
	require.True(t, maps.Equal(telemetryData, <-handler3.telemetry))
	require.True(t, maps.Equal(registrationData, <-handler3.registration))
}

func TestCacheStoreAddSameDevice(t *testing.T) {
	db := newTestCacheStoreDB()
	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	require.Nil(t, store.Add("http://foo.bar.com:123", "test-type", "foo-bar-com"))

	require.Equal(t, ErrDeviceExist,
		store.Add("http://foo.bar.com:123", "test-type", "foo-bar-com"))
}

func TestCacheStoreNoopDB(t *testing.T) {
	db := &stcore.NoopDB{}
	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	deviceURI := "http://foo.bar.com:123"
	deviceDesc := "foo-bar-com"
	deviceType := "test-type"

	require.Nil(t, store.Add(deviceURI, deviceType, deviceDesc))
	require.Nil(t, store.Remove(deviceURI))
}

func TestCacheStoreRestoreInvalidFormat(t *testing.T) {
	deviceURI := "http://foo.bar.com:123"
	deviceDesc := "foo-bar-com"

	db := newTestCacheStoreDB()
	require.Nil(t, db.Write(deviceURI, []byte(deviceDesc)))
	_, err := db.Read(deviceURI)
	require.Nil(t, err)

	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	_, err = db.Read(deviceURI)
	require.Equal(t, status.StatusNoData, err)
}

func TestCacheStoreRestoreUnsupportedScheme(t *testing.T) {
	deviceURI := "ftp://foo.bar.com:123"
	deviceDesc := "foo-bar-com"

	storageItem := StorageItem{
		Desc:      deviceDesc,
		Timestamp: time.Now().Unix(),
	}
	buf, err := storageItem.MarshalBinary()
	require.Nil(t, err)
	require.NotNil(t, buf)

	db := newTestCacheStoreDB()
	require.Nil(t, db.Write(deviceURI, buf))
	_, err = db.Read(deviceURI)
	require.Nil(t, err)

	clock := &testCacheStoreClock{}
	handler := newTestCacheStoreDataHandler()

	storeParams := CacheStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 100
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewCacheStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		sysnet.NewResolveStore(),
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Stop())
	}()

	_, err = db.Read(deviceURI)
	require.Equal(t, status.StatusNoData, err)
}
