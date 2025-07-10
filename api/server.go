package api

import (
	"context"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Server struct {
	router *http.ServeMux
}

func NewServer(driver *neo4j.DriverWithContext, dbCtx *context.Context) *Server {
	return &Server{
		router: setupMux(driver, dbCtx),
	}
}

func (s *Server) Start(port string) error {
	return http.ListenAndServe(port, s.router)
}
