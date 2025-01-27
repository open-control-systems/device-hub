package hthandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-control-systems/device-hub/components/http/htcore"
)

type testClock struct {
	mu        sync.Mutex
	timestamp int64
	setErr    error
	getErr    error
}

func (c *testClock) SetTimestamp(timestamp int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.setErr != nil {
		return c.setErr
	}

	c.timestamp = timestamp

	return nil
}

func (c *testClock) GetTimestamp() (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.getErr != nil {
		return -1, c.getErr
	}

	return c.timestamp, nil
}

func newTestClock(timestamp int64) *testClock {
	return &testClock{
		timestamp: timestamp,
	}
}

func TestSystemTimeHandlerSetGetTimestamp(t *testing.T) {
	currTimestamp := int64(123)
	startPoint := currTimestamp * 2

	testClock := newTestClock(currTimestamp)
	handler := NewSystemTimeHandler(testClock, time.Unix(startPoint, 0))

	mux := http.NewServeMux()
	mux.Handle("/api/v1/system/time", handler)

	server := httptest.NewServer(mux)
	defer server.Close()

	url := server.URL + "/api/v1/system/time"
	timeout := time.Second * 10
	ctx := context.Background()
	client := htcore.NewDefaultClient()

	clock := htcore.NewSystemClock(ctx, client, url, timeout)

	recvTimestamp, err := clock.GetTimestamp()
	require.Nil(t, err)
	require.Equal(t, int64(-1), recvTimestamp)

	newTimestamp := startPoint * 2
	require.NotEqual(t, currTimestamp, newTimestamp)

	require.Nil(t, clock.SetTimestamp(newTimestamp))

	recvTimestamp, err = clock.GetTimestamp()
	require.Nil(t, err)
	require.NotEqual(t, currTimestamp, recvTimestamp)
	require.Equal(t, newTimestamp, recvTimestamp)
}
