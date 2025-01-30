package httransport

import (
	"fmt"
	"net"
	"net/http"

	"github.com/open-control-systems/device-hub/components/system/sysnet"
)

// ResolveRoundTripper allows to perform hostname resolving before sending the HTTP request.
type ResolveRoundTripper struct {
	rs sysnet.Resolver
	rt http.RoundTripper
}

// NewResolveRoundTripper initializes round tripper.
//
// Parameters:
//   - rs to resolve HTTP addresses.
//   - rt to perform an actual HTTP transaction.
func NewResolveRoundTripper(rs sysnet.Resolver, rt http.RoundTripper) *ResolveRoundTripper {
	return &ResolveRoundTripper{
		rs: rs,
		rt: rt,
	}
}

// RoundTrip resolves HTTP address and perform HTTP transaction.
func (r *ResolveRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	addr, err := r.rs.Resolve(req.Context(), req.URL.Hostname())
	if err != nil {
		return nil, fmt.Errorf(
			"resolve-round-tripper: failed to resolve HTTP address: host=%s err=%v",
			req.URL.Host, err)
	}

	req.URL.Host = net.JoinHostPort(addr.String(), req.URL.Port())

	return r.rt.RoundTrip(req)
}
