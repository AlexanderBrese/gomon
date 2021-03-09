package browsersync

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AlexanderBrese/gomon/pkg/logging"
)

const route = "/sync"

// Server serves a REST route the client connects to receive sync messages
type Server struct {
	hub    *Hub
	srv    *http.Server
	logger *logging.Logger
}

// NewServer creates a new Server with the port provided
func NewServer(port int, l *logging.Logger) *Server {
	return &Server{
		hub:    NewHub(),
		srv:    &http.Server{Addr: fmt.Sprintf(":%d", port)},
		logger: l,
	}
}

// Start starts the server and lets the hub listen for clients
func (s *Server) Start() {
	s.startHub()
	s.setupRoute()
	s.startServer()
}

// Sync sends a sync message to the clients
func (s *Server) Sync() {
	message := bytes.TrimSpace([]byte("sync"))
	s.hub.broadcast <- message
}

// Stop stops the hub and the server gracefully
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.hub.stop()
	return s.srv.Shutdown(ctx)
}

func (s *Server) setupRoute() {
	http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		if err := communicate(s.hub, w, r); err != nil {
			s.logger.Main("error: failed to setup route: %s", err)
			return
		}
	})
}

func (s *Server) startHub() {
	go s.hub.listen()
}

func (s *Server) startServer() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			s.logger.Main("error: failed to serve sync server: %s", err)
			return
		}
		s.logger.Sync("Serving sync server at: %s", s.srv.Addr)
	}()
}
