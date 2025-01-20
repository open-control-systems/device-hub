package pipdevice

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-control-systems/device-hub/components/device"
	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/storage/stcore"
	"github.com/stretchr/testify/require"
)

type testPipelineStoreDB struct {
	data map[string]stcore.Blob
}

func newTestPipelineStoreDB() *testPipelineStoreDB {
	return &testPipelineStoreDB{
		data: make(map[string]stcore.Blob),
	}
}

func (d *testPipelineStoreDB) Read(key string) (stcore.Blob, error) {
	blob, ok := d.data[key]
	if !ok {
		return stcore.Blob{}, status.StatusNoData
	}

	return blob, nil
}

func (d *testPipelineStoreDB) Write(key string, blob stcore.Blob) error {
	b := stcore.Blob{}

	b.Data = make([]byte, len(blob.Data))
	copy(b.Data, blob.Data)

	d.data[key] = b

	return nil
}

func (d *testPipelineStoreDB) Remove(key string) error {
	delete(d.data, key)

	return nil
}

func (d *testPipelineStoreDB) ForEach(fn func(key string, b stcore.Blob) error) error {
	for k, v := range d.data {
		if err := fn(k, v); err != nil {
			return err
		}
	}

	return nil
}

func (d *testPipelineStoreDB) count() int {
	return len(d.data)
}

func (*testPipelineStoreDB) Close() error {
	return nil
}

type testPipelineStoreDataHandler struct {
	telemetry    chan device.JSON
	registration chan device.JSON
}

func newTestPipelineStoreDataHandler() *testPipelineStoreDataHandler {
	return &testPipelineStoreDataHandler{
		telemetry:    make(chan device.JSON),
		registration: make(chan device.JSON),
	}
}

func (h *testPipelineStoreDataHandler) HandleTelemetry(_ string, js device.JSON) error {
	select {
	case h.telemetry <- maps.Clone(js):
	default:
	}

	return nil
}

func (h *testPipelineStoreDataHandler) HandleRegistration(_ string, js device.JSON) error {
	select {
	case h.registration <- maps.Clone(js):
	default:
	}

	return nil
}

type testPipelineStoreClock struct {
	timestamp int64
}

func (c *testPipelineStoreClock) SetTimestamp(timestamp int64) error {
	c.timestamp = timestamp

	return nil
}

func (c *testPipelineStoreClock) GetTimestamp() (int64, error) {
	return c.timestamp, nil
}

type testPipelineStoreHTTPDataHandler struct {
	js device.JSON
}

func newTestPipelineStoreHTTPDataHandler(data device.JSON) *testPipelineStoreHTTPDataHandler {
	return &testPipelineStoreHTTPDataHandler{
		js: maps.Clone(data),
	}
}

func (h *testPipelineStoreHTTPDataHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(h.js); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TestPipelineStoreStartCloseEmpty(t *testing.T) {
	db := newTestPipelineStoreDB()
	clock := &testPipelineStoreClock{}
	handler := newTestPipelineStoreDataHandler()

	storeParams := PipelineStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 500
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewPipelineStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Close())
	}()

	store.Start()
}

func TestPipelineStoreCloseNoStart(t *testing.T) {
	db := newTestPipelineStoreDB()
	clock := &testPipelineStoreClock{}
	handler := newTestPipelineStoreDataHandler()

	storeParams := PipelineStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 500
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewPipelineStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Close())
	}()
}

func TestPipelineStoreGetDescEmpty(t *testing.T) {
	db := newTestPipelineStoreDB()
	clock := &testPipelineStoreClock{}
	handler := newTestPipelineStoreDataHandler()

	storeParams := PipelineStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 500
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewPipelineStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Close())
	}()

	descs := store.GetDesc()
	require.Empty(t, descs)
}

func TestPipelineStoreRemoveNoAdd(t *testing.T) {
	db := newTestPipelineStoreDB()
	clock := &testPipelineStoreClock{}
	handler := newTestPipelineStoreDataHandler()

	storeParams := PipelineStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 500
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewPipelineStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Close())
	}()

	require.Equal(t, status.StatusNoData, store.Remove("foo-bar-baz"))
}

func TestPipelineStoreAddURIUnsupportedScheme(t *testing.T) {
	db := newTestPipelineStoreDB()
	clock := &testPipelineStoreClock{}
	handler := newTestPipelineStoreDataHandler()

	storeParams := PipelineStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 500
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewPipelineStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Close())
	}()

	require.Equal(t, status.StatusNotSupported, store.Add("foo-bar-baz", "foo-bar-baz"))
}

func TestPipelineStoreAddRemoveResourceNoResponse(t *testing.T) {
	db := newTestPipelineStoreDB()
	clock := &testPipelineStoreClock{}
	handler := newTestPipelineStoreDataHandler()

	storeParams := PipelineStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 500
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	ctx, cancelFunc := context.WithTimeoutCause(
		context.Background(),
		time.Second*2,
		status.StatusTimeout,
	)
	defer cancelFunc()

	store := NewPipelineStore(
		ctx,
		clock,
		clock,
		handler,
		db,
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Close())
	}()

	tests := []struct {
		uri  string
		desc string
	}{
		{"http://device.example.com/api/v10", "foo-bar-baz"},
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

func TestPipelineStoreAddRemove(t *testing.T) {
	db := newTestPipelineStoreDB()
	clock := &testPipelineStoreClock{}
	handler := newTestPipelineStoreDataHandler()

	storeParams := PipelineStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 500
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewPipelineStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Close())
	}()

	deviceID := "0xABCD"

	telemetryData := make(device.JSON)
	telemetryData["timestamp"] = float64(123)
	telemetryData["temperature"] = float64(123.222)

	registrationData := make(device.JSON)
	registrationData["timestamp"] = float64(123)
	registrationData["device_id"] = deviceID

	telemetryHandler := newTestPipelineStoreHTTPDataHandler(telemetryData)
	registrationHandler := newTestPipelineStoreHTTPDataHandler(registrationData)

	mux := http.NewServeMux()
	mux.Handle("/telemetry", telemetryHandler)
	mux.Handle("/registration", registrationHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	require.Nil(t, store.Add(server.URL, "foo-bar-baz"))

	require.True(t, maps.Equal(telemetryData, <-handler.telemetry))
	require.True(t, maps.Equal(registrationData, <-handler.registration))
}

func TestPipelineStoreRestore(t *testing.T) {
	db := newTestPipelineStoreDB()

	makeStore := func(d stcore.DB, h device.DataHandler) *PipelineStore {
		clock := &testPipelineStoreClock{}

		storeParams := PipelineStoreParams{}
		storeParams.HTTP.FetchInterval = time.Millisecond * 500
		storeParams.HTTP.FetchTimeout = time.Millisecond * 100

		return NewPipelineStore(
			context.Background(),
			clock,
			clock,
			h,
			d,
			storeParams,
		)
	}

	handler1 := newTestPipelineStoreDataHandler()
	store1 := makeStore(db, handler1)

	require.Empty(t, store1.GetDesc())

	deviceID := "0xABCD"

	telemetryData := make(device.JSON)
	telemetryData["timestamp"] = float64(123)
	telemetryData["temperature"] = float64(123.222)

	registrationData := make(device.JSON)
	registrationData["timestamp"] = float64(123)
	registrationData["device_id"] = deviceID

	telemetryHandler := newTestPipelineStoreHTTPDataHandler(telemetryData)
	registrationHandler := newTestPipelineStoreHTTPDataHandler(registrationData)

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

	require.Nil(t, store1.Close())

	handler2 := newTestPipelineStoreDataHandler()
	store2 := makeStore(db, handler2)

	descs := store2.GetDesc()
	require.Equal(t, 1, len(descs))

	desc := descs[0]
	require.Equal(t, deviceURI, desc.URI)
	require.Equal(t, deviceDesc, desc.Desc)

	store2.Start()

	require.NotNil(t, store2.Add(deviceURI, deviceDesc))
	require.True(t, maps.Equal(telemetryData, <-handler2.telemetry))
	require.True(t, maps.Equal(registrationData, <-handler2.registration))

	require.Nil(t, store2.Remove(deviceURI))

	handler3 := newTestPipelineStoreDataHandler()
	store3 := makeStore(db, handler3)

	require.Nil(t, store3.Add(deviceURI, deviceDesc))
	require.True(t, maps.Equal(telemetryData, <-handler3.telemetry))
	require.True(t, maps.Equal(registrationData, <-handler3.registration))
}

func TestPipelineStoreAddSameDevice(t *testing.T) {
	db := newTestPipelineStoreDB()
	clock := &testPipelineStoreClock{}
	handler := newTestPipelineStoreDataHandler()

	storeParams := PipelineStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 500
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewPipelineStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Close())
	}()

	require.Nil(t, store.Add("http://foo.bar.com", "foo-bar-com"))
	require.NotNil(t, store.Add("http://foo.bar.com", "foo-bar-com"))
}

func TestPipelineStoreNoopDB(t *testing.T) {
	db := &stcore.NoopDB{}
	clock := &testPipelineStoreClock{}
	handler := newTestPipelineStoreDataHandler()

	storeParams := PipelineStoreParams{}
	storeParams.HTTP.FetchInterval = time.Millisecond * 500
	storeParams.HTTP.FetchTimeout = time.Millisecond * 100

	store := NewPipelineStore(
		context.Background(),
		clock,
		clock,
		handler,
		db,
		storeParams,
	)
	defer func() {
		require.Nil(t, store.Close())
	}()

	deviceURI := "http://foo.bar.com"
	deviceDesc := "foo-bar-com"

	require.Nil(t, store.Add(deviceURI, deviceDesc))
	require.Nil(t, store.Remove(deviceURI))
}
