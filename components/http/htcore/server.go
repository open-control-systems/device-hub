package htcore

import (
	"net"
	"net/http"
	"strconv"

	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// Server is a wrapper for http.Server.
type Server struct {
	server http.Server
	ln     net.Listener
	doneCh chan struct{}
	url    string
	port   int
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
		port:   params.Port,
	}, nil
}

// Start runs the server.
func (s *Server) Start() error {
	go s.run()

	return nil
}

// Stop stops the server and waits until it finishes.
func (s *Server) Stop() error {
	err := s.server.Close()

	_ = s.ln.Close()

	<-s.doneCh

	return err
}

// URL returns base URL of form http://ipaddr:port with no trailing slash.
func (s *Server) URL() string {
	return s.url
}

// Port returns the port to which the server socket is bound.
func (s *Server) Port() int {
	return s.port
}

func (s *Server) run() {
	defer close(s.doneCh)

	if err := s.server.Serve(s.ln); err != nil && err != http.ErrServerClosed {
		syscore.LogErr.Printf("failed to serve connection: %v", err)
	}
}
