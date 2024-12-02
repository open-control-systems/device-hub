package htclient

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/open-control-systems/device-hub/components/http/httransport"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
)

// Standard HTTP client wrapper to simplify response reading.
type HttpClient struct {
	http.Client
}

// General purpose HTTP client.
func NewDefaultClient() *HttpClient {
	return &HttpClient{}
}

// HTTP client with customer resolving rules.
func NewResolveClient(resolver sysnet.Resolver) *HttpClient {
	return &HttpClient{
		Client: http.Client{
			Transport: httransport.NewResolveRoundTripper(resolver, http.DefaultTransport),
		},
	}
}

// Do sends a request, receives a response, and fully reads the response body.
func (c *HttpClient) Do(req *http.Request) (*http.Response, []byte, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var body []byte
	switch resp.ContentLength {
	case -1:
		body, err = ioutil.ReadAll(resp.Body)
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
