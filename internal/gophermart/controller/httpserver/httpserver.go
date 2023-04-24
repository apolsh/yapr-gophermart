package httpserver

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"
)

const (
	shutdownTimeout        = 3 * time.Second
	authorizationHeaderKey = "Authorization"
)

type Server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

func NewServer(server *http.Server) *Server {

	s := &Server{
		server:          server,
		notify:          make(chan error, 1),
		shutdownTimeout: shutdownTimeout,
	}

	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) StartTLS(cfg *tls.Config) error {
	s.server.TLSConfig = cfg
	return s.server.ListenAndServeTLS("", "")
}

func (s *Server) Stop(ctx context.Context) error {
	s.server.SetKeepAlivesEnabled(false)
	return s.server.Shutdown(ctx)
}
