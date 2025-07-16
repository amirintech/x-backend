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

var chainMiddleware = chain(corsMiddleware, authMiddleware)

func setupMux(db *neo4j.DriverWithContext, dbCtx *context.Context, authConfig *oauth2.Config) *http.ServeMux {
	router := http.NewServeMux()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Setup services
	notificationsService := services.NewNotificationsService()
	feedService := services.NewFeedService()

	// User routes
	userStore := stores.NewUserStore(db, dbCtx, notificationsService)
	userHandlers := handlers.NewUserHandlers(&userStore)
	setupUserRoutes(router, userHandlers)

	// Auth routes
	authHandlers := handlers.NewAuthHandlers(&userStore)
	setupAuthRoutes(router, authHandlers, authConfig)

	// Tweet routes
	tweetStore := stores.NewTweetStore(db, dbCtx, notificationsService, feedService)
	tweetHandlers := handlers.NewTweetHandlers(&tweetStore)
	setupTweetRoutes(router, tweetHandlers)

	// Notifications routes
	notificationsStore := stores.NewNotificationsStore(db, dbCtx)
	notificationsHandlers := handlers.NewNotificationsHandlers(notificationsService, notificationsStore)
	setupNotificationsRoutes(router, notificationsHandlers)

	// Feed routes
	feedHandlers := handlers.NewFeedHandlers(feedService)
	setupFeedRoutes(router, feedHandlers)

	return router
}

func setupAuthRoutes(router *http.ServeMux, authHandlers *handlers.AuthHandlers, authConfig *oauth2.Config) {
	router.HandleFunc("POST /api/auth/login", corsMiddleware(authHandlers.Login))
	router.HandleFunc("POST /api/auth/register", corsMiddleware(authHandlers.Register))
	router.HandleFunc("GET /api/oauth/login", corsMiddleware(authHandlers.OAuthLoginHandler(authConfig)))
	router.HandleFunc("GET /api/oauth/callback", corsMiddleware(authHandlers.OAuthCallbackHandler(authConfig)))
	router.HandleFunc("POST /api/auth/logout", corsMiddleware(authHandlers.Logout))
}

func setupUserRoutes(router *http.ServeMux, userHandlers *handlers.UserHandlers) {
	router.HandleFunc("GET /api/users/id/{id}", corsMiddleware(userHandlers.GetUserByID))
	router.HandleFunc("GET /api/users/username/{username}", corsMiddleware(userHandlers.GetUserByUsername))
	router.HandleFunc("GET /api/users", chainMiddleware(userHandlers.GetCurrentUser))
	router.HandleFunc("PUT /api/users", chainMiddleware(userHandlers.UpdateUser))
}

func setupTweetRoutes(router *http.ServeMux, tweetHandlers *handlers.TweetHandlers) {
	router.HandleFunc("GET /api/tweets", chainMiddleware(tweetHandlers.GetUsersWithTweets))
	router.HandleFunc("GET /api/tweets/{id}", chainMiddleware(tweetHandlers.GetTweetByID))
	router.HandleFunc("POST /api/tweets", chainMiddleware(tweetHandlers.CreateTweet))
	router.HandleFunc("POST /api/tweets/{id}/like", chainMiddleware(tweetHandlers.LikeTweet))
	router.HandleFunc("POST /api/tweets/{id}/unlike", chainMiddleware(tweetHandlers.UnlikeTweet))
}

func setupNotificationsRoutes(router *http.ServeMux, notificationsHandlers *handlers.NotificationsHandlers) {
	router.HandleFunc("GET /api/notifications/{type}/{userID}", notificationsHandlers.StreamNotifications)
}

func setupFeedRoutes(router *http.ServeMux, feedHandlers *handlers.FeedHandlers) {
	router.HandleFunc("GET /api/feed", authMiddleware(feedHandlers.StreamFeed))
}
