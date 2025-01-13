package htcore

import (
	"io"
	"net/http"

	"github.com/open-control-systems/device-hub/components/http/httransport"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
)

// HTTPClient is a standard HTTP client wrapper that simplifies reading responses.
type HTTPClient struct {
	http.Client
}

// NewDefaultClient creates a general purpose HTTP client.
func NewDefaultClient() *HTTPClient {
	return &HTTPClient{}
}

// NewResolveClient creates HTTP client with custom resolving rules.
func NewResolveClient(resolver sysnet.Resolver) *HTTPClient {
	return &HTTPClient{
		Client: http.Client{
			Transport: httransport.NewResolveRoundTripper(resolver, http.DefaultTransport),
		},
	}
}

// Do sends a request, receives a response, and fully reads the response body.
func (c *HTTPClient) Do(req *http.Request) (*http.Response, []byte, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var body []byte
	switch resp.ContentLength {
	case -1:
		body, err = io.ReadAll(resp.Body)
	case 0:
		body, err = []byte{}, nil
	default:
		body = make([]byte, resp.ContentLength)
		_, err = io.ReadFull(resp.Body, body)
	}
	if err != nil {
		return nil, nil, err
	}

	return resp, body, nil
}
