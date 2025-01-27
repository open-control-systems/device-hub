package htcore

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// SystemClock handles the UNIX time for the HTTP resource.
type SystemClock struct {
	url     string
	timeout time.Duration
	ctx     context.Context
	client  *HTTPClient
}

// NewSystemClock initializes SystemClock.
//
// Parameters:
//   - ctx to pass to the HTTP request.
//   - client to perform an actual HTTP request.
//   - url - HTTP URL.
//   - timeout - HTTP request timeout.
func NewSystemClock(
	ctx context.Context,
	client *HTTPClient,
	url string,
	timeout time.Duration,
) *SystemClock {
	return &SystemClock{
		url:     url,
		timeout: timeout,
		ctx:     ctx,
		client:  client,
	}
}

// SetTimestamp sets the UNIX time for a remoute resource.
func (c *SystemClock) SetTimestamp(timestamp int64) error {
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
	if err != nil {
		return err
	}

	query := req.URL.Query()
	query.Set("value", strconv.FormatInt(timestamp, 10))
	req.URL.RawQuery = query.Encode()

	resp, _, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http-system-clock: failed to send time: code=%v", resp.StatusCode)
	}

	return nil
}

// GetTimestamp gets the UNIX time from a remote resource.
func (c *SystemClock) GetTimestamp() (int64, error) {
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
	if err != nil {
		return -1, err
	}

	resp, body, err := c.client.Do(req)
	if err != nil {
		return -1, err
	}

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("http-system-clock: failed to receive time: code=%v",
			resp.StatusCode)
	}

	timestamp, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return -1, err
	}

	return timestamp, nil
}
