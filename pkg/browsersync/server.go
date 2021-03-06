package browsersync

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

const (
	ROUTE = "/sync"
)

type Server struct {
	hub *Hub
}

func NewServer() *Server {
	return &Server{
		hub: NewHub(),
	}
}

func (s *Server) Start(port int) {
	s.startHub()
	s.setupRoute()
	s.startServer(port)
}

func (s *Server) Sync() {
	message := bytes.TrimSpace([]byte("sync"))
	s.hub.broadcast <- message
}

func (s *Server) startHub() {
	go s.hub.listen()
}

func (s *Server) setupRoute() {
	http.HandleFunc(ROUTE, func(w http.ResponseWriter, r *http.Request) {
		communicate(s.hub, w, r)
	})
}

func (s *Server) startServer(port int) {
	go s.serve(port)
	log.Println("Serving sync server at", port)
}

func (s *Server) serve(port int) error {
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
