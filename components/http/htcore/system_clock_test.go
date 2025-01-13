package htcore

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testHandler struct {
	timestamp int64
	err       error
}

func newTestHandler(timestamp int64) *testHandler {
	return &testHandler{
		timestamp: timestamp,
	}
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.err != nil {
		http.Error(w, h.err.Error(), http.StatusInternalServerError)

		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "error: unsupported method", http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	response := ""

	str := r.URL.Query().Get("value")
	if str == "" {
		response = strconv.FormatInt(h.timestamp, 10)
	} else {
		timestamp, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		h.timestamp = timestamp

		response = "OK"
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.WriteHeader(http.StatusOK)

	if _, err := fmt.Fprint(w, response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TestHTTPSystemClockSetGetTimestamp(t *testing.T) {
	currTimestamp := int64(-1)

	handler := newTestHandler(currTimestamp)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/system/time", handler)

	server := httptest.NewServer(mux)
	defer server.Close()

	url := server.URL + "/api/v1/system/time"
	timeout := time.Second * 10
	ctx := context.Background()
	client := NewDefaultClient()

	clock := NewSystemClock(ctx, client, url, timeout)

	recvTimestamp, err := clock.GetTimestamp()
	require.Nil(t, err)
	require.Equal(t, currTimestamp, recvTimestamp)

	newTimestamp := time.Now().Unix()
	require.NotEqual(t, currTimestamp, newTimestamp)

	require.Nil(t, clock.SetTimestamp(newTimestamp))

	recvTimestamp, err = clock.GetTimestamp()
	require.Nil(t, err)
	require.NotEqual(t, currTimestamp, recvTimestamp)
	require.Equal(t, newTimestamp, recvTimestamp)
}
