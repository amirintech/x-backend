package api

import (
	"context"
	"net/http"

	"github.com/aimrintech/x-backend/handlers"
	"github.com/aimrintech/x-backend/stores"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func  setupMux(db *neo4j.DriverWithContext, dbCtx *context.Context) *http.ServeMux {
	router := http.NewServeMux()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// User routes
	userStore := stores.NewUserStore(db, dbCtx)
	userHandlers := handlers.NewUserHandlers(&userStore)
	setupUserRoutes(router, userHandlers)

	// Auth routes
	authHandlers := handlers.NewAuthHandlers(&userStore)
	setupAuthRoutes(router, authHandlers)

	// Tweet routes
	tweetStore := stores.NewTweetStore(db, dbCtx)
	tweetHandlers := handlers.NewTweetHandlers(&tweetStore)
	setupTweetRoutes(router, tweetHandlers)

	return router		
}

func setupAuthRoutes(router *http.ServeMux, authHandlers *handlers.AuthHandlers) {
	router.HandleFunc("POST /api/auth/login", authHandlers.Login)
	router.HandleFunc("POST /api/auth/register", authHandlers.Register)
}

func setupUserRoutes(router *http.ServeMux, userHandlers *handlers.UserHandlers) {
	router.HandleFunc("GET /api/users/{id}", userHandlers.GetUserByID)
	registerProtectedRoute(router, userHandlers.GetCurrentUser, "GET /api/users")
	registerProtectedRoute(router, userHandlers.UpdateUser, "PUT /api/users")
}

func setupTweetRoutes(router *http.ServeMux, tweetHandlers *handlers.TweetHandlers) {

}

func registerProtectedRoute(router *http.ServeMux, handler http.HandlerFunc, path string) {
	router.HandleFunc(path, jwtAuthMiddleware(handler))
}

