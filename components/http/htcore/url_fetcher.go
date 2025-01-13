package htcore

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// URLFetcher sends requests to the configured HTTP endpoint.
type URLFetcher struct {
	ctx     context.Context
	url     string
	timeout time.Duration
	client  *HTTPClient
}

// NewURLFetcher initializes URL Fetcher.
//
// Parameters:
//   - ctx to pass to the HTTP request.
//   - client to perform an actual HTTP request.
//   - url - HTTP URL.
//   - timeout - HTTP request timeout.
func NewURLFetcher(
	ctx context.Context,
	client *HTTPClient,
	url string,
	timeout time.Duration,
) *URLFetcher {
	return &URLFetcher{
		ctx:     ctx,
		url:     url,
		timeout: timeout,
		client:  client,
	}
}

// Fetch data from the HTTP resource.
func (f *URLFetcher) Fetch() ([]byte, error) {
	ctx, cancel := context.WithTimeout(f.ctx, f.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", f.url, nil)
	if err != nil {
		return nil, err
	}

	resp, body, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("url-fetcher: failed to fetch data: code=%v", resp.StatusCode)
	}

	return body, nil
}
