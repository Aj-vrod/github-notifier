package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Server represents the HTTP server for the github-notifier API
type Server struct {
	router *mux.Router
	port   int
}

// NewServer creates a new API server instance
// port specifies which port the server will listen on
func NewServer(port int) *Server {
	s := &Server{
		router: mux.NewRouter(),
		port:   port,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes for the API
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/health", HandleHealth).Methods("GET")
	s.router.HandleFunc("/api/v1/subscribe", HandleSubscribe).Methods("POST")
}

// Start begins listening for HTTP requests on the configured port
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server on port %d", s.port)
	return http.ListenAndServe(addr, s.router)
}

// GetRouter returns the underlying mux router
// This is useful for testing
func (s *Server) GetRouter() *mux.Router {
	return s.router
}
