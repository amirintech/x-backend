package api

import (
	"context"
	"net/http"

	"github.com/aimrintech/x-backend/handlers"
	"github.com/aimrintech/x-backend/services"
	"github.com/aimrintech/x-backend/stores"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/oauth2"
)

var chainMiddleware = chain(headersMiddleware, authMiddleware)

func setupMux(db *neo4j.DriverWithContext, dbCtx *context.Context, authConfig *oauth2.Config) *http.ServeMux {
	router := http.NewServeMux()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Setup services
	notificationsService := services.NewNotificationsService()

	// User routes
	userStore := stores.NewUserStore(db, dbCtx, notificationsService)
	userHandlers := handlers.NewUserHandlers(&userStore)
	setupUserRoutes(router, userHandlers)

	// Auth routes
	authHandlers := handlers.NewAuthHandlers(&userStore)
	setupAuthRoutes(router, authHandlers, authConfig)

	// Tweet routes
	tweetStore := stores.NewTweetStore(db, dbCtx, notificationsService)
	tweetHandlers := handlers.NewTweetHandlers(&tweetStore)
	setupTweetRoutes(router, tweetHandlers)

	// Notifications routes
	notificationsStore := stores.NewNotificationsStore(db, dbCtx)
	notificationsHandlers := handlers.NewNotificationsHandlers(notificationsService, notificationsStore)
	setupNotificationsRoutes(router, notificationsHandlers)

	return router
}

func setupAuthRoutes(router *http.ServeMux, authHandlers *handlers.AuthHandlers, authConfig *oauth2.Config) {
	router.HandleFunc("POST /api/auth/login", headersMiddleware(authHandlers.Login))
	router.HandleFunc("POST /api/auth/register", headersMiddleware(authHandlers.Register))
	router.HandleFunc("GET /api/oauth/login", headersMiddleware(authHandlers.OAuthLoginHandler(authConfig)))
	router.HandleFunc("GET /api/oauth/callback", headersMiddleware(authHandlers.OAuthCallbackHandler(authConfig)))
	router.HandleFunc("POST /api/auth/logout", headersMiddleware(authHandlers.Logout))
}

func setupUserRoutes(router *http.ServeMux, userHandlers *handlers.UserHandlers) {
	router.HandleFunc("GET /api/users/{id}", headersMiddleware(userHandlers.GetUserByID))
	router.HandleFunc("GET /api/users", chainMiddleware(userHandlers.GetCurrentUser))
	router.HandleFunc("PUT /api/users", chainMiddleware(userHandlers.UpdateUser))
}

func setupTweetRoutes(router *http.ServeMux, tweetHandlers *handlers.TweetHandlers) {
	router.HandleFunc("GET /api/tweets", chainMiddleware(tweetHandlers.GetTweets))
	router.HandleFunc("GET /api/tweets/{id}", chainMiddleware(tweetHandlers.GetTweetByID))
	router.HandleFunc("POST /api/tweets", chainMiddleware(tweetHandlers.CreateTweet))
	router.HandleFunc("POST /api/tweets/{id}/like", chainMiddleware(tweetHandlers.LikeTweet))
	router.HandleFunc("POST /api/tweets/{id}/unlike", chainMiddleware(tweetHandlers.UnlikeTweet))

}

func setupNotificationsRoutes(router *http.ServeMux, notificationsHandlers *handlers.NotificationsHandlers) {
	router.HandleFunc("GET /api/notifications/{type}/{userID}", notificationsHandlers.StreamNotifications)
}
