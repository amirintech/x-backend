package api

import (
	"context"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/oauth2"
)

type Server struct {
	router *http.ServeMux
}

func NewServer(driver *neo4j.DriverWithContext, dbCtx *context.Context, authConfig *oauth2.Config) *Server {
	return &Server{
		router: setupMux(driver, dbCtx, authConfig),
	}
}

func (s *Server) Start(port string) error {
	return http.ListenAndServe(port, s.router)
}
