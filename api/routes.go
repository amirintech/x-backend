package api

import (
	"context"
	"net/http"

	"github.com/aimrintech/x-backend/handlers"
	"github.com/aimrintech/x-backend/services/notifications"
	"github.com/aimrintech/x-backend/stores"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var chainMiddleware = chain(headersMiddleware, authMiddleware)

func setupMux(db *neo4j.DriverWithContext, dbCtx *context.Context) *http.ServeMux {
	router := http.NewServeMux()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Setup services
	notificationsService := notifications.NewNotificationsService()

	// HTML FOR TESTING
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/index.html")
	})

	// User routes
	userStore := stores.NewUserStore(db, dbCtx, notificationsService)
	userHandlers := handlers.NewUserHandlers(&userStore)
	setupUserRoutes(router, userHandlers)

	// Auth routes
	authHandlers := handlers.NewAuthHandlers(&userStore)
	setupAuthRoutes(router, authHandlers)

	// Tweet routes
	tweetStore := stores.NewTweetStore(db, dbCtx, notificationsService)
	tweetHandlers := handlers.NewTweetHandlers(&tweetStore)
	setupTweetRoutes(router, tweetHandlers)

	// Notifications routes
	notificationsHandlers := handlers.NewNotificationsHandlers(notificationsService)
	setupNotificationsRoutes(router, notificationsHandlers)

	return router
}

func setupAuthRoutes(router *http.ServeMux, authHandlers *handlers.AuthHandlers) {
	router.HandleFunc("POST /api/auth/login", headersMiddleware(authHandlers.Login))
	router.HandleFunc("POST /api/auth/register", headersMiddleware(authHandlers.Register))
}

func setupUserRoutes(router *http.ServeMux, userHandlers *handlers.UserHandlers) {
	router.HandleFunc("GET /api/users/{id}", headersMiddleware(userHandlers.GetUserByID))
	router.HandleFunc("GET /api/users", chainMiddleware(userHandlers.GetCurrentUser))
	router.HandleFunc("PUT /api/users", chainMiddleware(userHandlers.UpdateUser))
}

func setupTweetRoutes(router *http.ServeMux, tweetHandlers *handlers.TweetHandlers) {

}

func setupNotificationsRoutes(router *http.ServeMux, notificationsHandlers *handlers.NotificationsHandlers) {
	router.HandleFunc("GET /api/notifications", chainMiddleware(notificationsHandlers.StreamNotifications))
}
