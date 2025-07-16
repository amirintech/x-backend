package utils

import (
	"net/http"

	"github.com/rs/cors"
)

var corsConfig = cors.New(cors.Options{
	AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowedHeaders:   []string{"Content-Type", "Authorization"},
	AllowCredentials: true,
})

func SetCORSHeaders(w http.ResponseWriter, r *http.Request) {
	corsConfig.ServeHTTP(w, r, func(w http.ResponseWriter, r *http.Request) {
	})
}

func HandleCORSPreflight(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodOptions {
		SetCORSHeaders(w, r)
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}
