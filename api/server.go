package api

import (
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Server struct {
	router *http.ServeMux
}

func NewServer(driver *neo4j.DriverWithContext) *Server {	
	return &Server{
		router: setupMux(driver),	
	}
}

func (s *Server) Start(port string) error {
	return http.ListenAndServe(port, s.router)
}