package htcore

import (
	"net"
	"net/http"
	"strconv"

	"github.com/open-control-systems/device-hub/components/core"
)

// Server is a wrapper for http.Server.
type Server struct {
	server http.Server
	ln     net.Listener
	doneCh chan struct{}
	url    string
}

// ServerParams contains server parameters.
type ServerParams struct {
	Host string
	Port int
}

// NewServer creates a new server.
//
// Notes:
//   - The server is not started.
//   - If host is empty, "0.0.0.0" is used.
//   - If port is zero, a random free port is chosen.
//
// References:
//   - The implementation is based on the httptest.Server.
func NewServer(handler http.Handler, params ServerParams) (*Server, error) {
	if params.Host == "" {
		params.Host = "0.0.0.0"
	}

	addr, err := net.ResolveTCPAddr("tcp", params.Host+":"+strconv.Itoa(params.Port))
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP(addr.Network(), addr)
	if err != nil {
		return nil, err
	}

	if params.Port == 0 {
		params.Port = ln.Addr().(*net.TCPAddr).Port
	}

	return &Server{
		server: http.Server{
			Addr:    addr.String(),
			Handler: handler,
		},
		ln:     ln,
		doneCh: make(chan struct{}),
		url:    "http://" + ln.Addr().String(),
	}, nil
}

// Start runs the server.
func (s *Server) Start() {
	go s.run()
}

// Close stops the server and waits until it finishes.
func (s *Server) Close() error {
	err := s.server.Close()

	_ = s.ln.Close()

	<-s.doneCh

	return err
}

// URL returns base URL of form http://ipaddr:port with no trailing slash.
func (s *Server) URL() string {
	return s.url
}

func (s *Server) run() {
	defer close(s.doneCh)

	if err := s.server.Serve(s.ln); err != nil {
		core.LogErr.Printf("http-server: failed to serve connection: %v\n", err)
	}
}
