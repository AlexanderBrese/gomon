package browsersync

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	ROUTE = "/sync"
)

type Server struct {
	hub *Hub
	srv *http.Server
}

type RouteHandler struct {
	hub *Hub
}

func NewServer(port int) *Server {
	return &Server{
		hub: NewHub(),
		srv: &http.Server{Addr: fmt.Sprintf(":%d", port)},
	}
}

func (s *Server) Start() {
	s.startHub()
	s.setupRoute()
	s.startServer()
}

func (s *Server) Sync() {
	message := bytes.TrimSpace([]byte("sync"))
	s.hub.broadcast <- message
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.hub.stop()
	return s.srv.Shutdown(ctx)
}

func (s *Server) setupRoute() {
	http.HandleFunc(ROUTE, func(w http.ResponseWriter, r *http.Request) {
		communicate(s.hub, w, r)
	})
}

func (s *Server) startHub() {
	go s.hub.listen()
}

func (s *Server) startServer() {
	go s.srv.ListenAndServe()
	log.Println("Serving sync server at", s.srv.Addr)
}
