package browsersync

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

const route = "/sync"

// Server serves a REST route the client connects to receive sync messages
type Server struct {
	hub *Hub
	srv *http.Server
}

// NewServer creates a new Server with the port provided
func NewServer(port int) *Server {
	return &Server{
		hub: NewHub(),
		srv: &http.Server{Addr: fmt.Sprintf(":%d", port)},
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
			// TODO: log
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
			// TODO: log
			return
		}
		log.Println("Serving sync server at", s.srv.Addr)
	}()
}
