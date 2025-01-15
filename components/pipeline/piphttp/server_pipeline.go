package piphttp

import (
	"net/http"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/http/htcore"
)

// ServerPipeline contains various building blocks for HTTP API.
type ServerPipeline struct {
	server *htcore.Server
	mux    *http.ServeMux
}

// NewServerPipeline initializes all components associated with the HTTP server.
//
// Parameters:
//   - closer - to register handlers for the underlying resource deallocation.
//   - serverParams - various HTTP server configuration parameters.
func NewServerPipeline(
	closer *core.FanoutCloser,
	serverParams htcore.ServerParams,
) (*ServerPipeline, error) {
	mux := http.NewServeMux()

	server, err := htcore.NewServer(mux, serverParams)
	if err != nil {
		return nil, err
	}
	closer.Add("http-server", server)

	core.LogInf.Printf("http-server-pipeline: starting HTTP server: URL=%s",
		server.URL())

	return &ServerPipeline{
		server: server,
		mux:    mux,
	}, nil
}

// GetServeMux returns the component to register HTTP endpoints.
func (p *ServerPipeline) GetServeMux() *http.ServeMux {
	return p.mux
}

// Start starts serving HTTP requests.
func (p *ServerPipeline) Start() {
	p.server.Start()
}
