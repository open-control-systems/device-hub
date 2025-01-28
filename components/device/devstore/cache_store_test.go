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
	data map[string]stcore.Blob
}

func newTestCacheStoreDB() *testCacheStoreDB {
	return &testCacheStoreDB{
		data: make(map[string]stcore.Blob),
	}
}

func (d *testCacheStoreDB) Read(key string) (stcore.Blob, error) {
	blob, ok := d.data[key]
	if !ok {
		return stcore.Blob{}, status.StatusNoData
	}

	return blob, nil
}

func (d *testCacheStoreDB) Write(key string, blob stcore.Blob) error {
	b := stcore.Blob{}

	b.Data = make([]byte, len(blob.Data))
	copy(b.Data, blob.Data)

	d.data[key] = b

	return nil
}

func (d *testCacheStoreDB) Remove(key string) error {
	delete(d.data, key)

	return nil
}

func (d *testCacheStoreDB) ForEach(fn func(key string, b stcore.Blob) error) error {
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

	require.Equal(t, status.StatusNotSupported, store.Add("foo-bar-baz", "foo-bar-baz"))
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
		desc string
	}{
		{"http://devcore.example.com/api/v10", "foo-bar-baz"},
		{"http://192.1.2.3:8787/api/v3", "foo-bar-baz"},
		{"https://192.1.2.3:1234", "foo-bar-baz"},
		{"http://bonsai-growlab.local/api/v1", "foo-bar-baz"},
	}

	for _, test := range tests {
		require.Nil(t, store.Add(test.uri, test.desc))
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

	require.Nil(t, store.Add(server.URL, "foo-bar-baz"))

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

	require.Nil(t, store1.Add(deviceURI, deviceDesc))

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

	require.Nil(t, store2.Start())

	require.NotNil(t, store2.Add(deviceURI, deviceDesc))
	require.True(t, maps.Equal(telemetryData, <-handler2.telemetry))
	require.True(t, maps.Equal(registrationData, <-handler2.registration))

	require.Nil(t, store2.Remove(deviceURI))

	handler3 := newTestCacheStoreDataHandler()
	store3 := makeStore(db, handler3)

	require.Nil(t, store3.Add(deviceURI, deviceDesc))
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

	require.Nil(t, store.Add("http://foo.bar.com", "foo-bar-com"))
	require.Equal(t, ErrDeviceExist, store.Add("http://foo.bar.com", "foo-bar-com"))
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

	deviceURI := "http://foo.bar.com"
	deviceDesc := "foo-bar-com"

	require.Nil(t, store.Add(deviceURI, deviceDesc))
	require.Nil(t, store.Remove(deviceURI))
}
