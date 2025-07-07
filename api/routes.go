package api

import (
	"net/http"

	"github.com/aimrintech/x-backend/handlers"
	"github.com/aimrintech/x-backend/stores"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)



func  setupMux(db *neo4j.DriverWithContext) *http.ServeMux {
	router := http.NewServeMux()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// User routes
	userStore := stores.NewUserStore(db)
	userHandlers := handlers.NewUserHandlers(&userStore)
	setupUserRoutes(router, userHandlers)

	// Tweet routes
	tweetStore := stores.NewTweetStore(db)
	tweetHandlers := handlers.NewTweetHandlers(&tweetStore)
	setupTweetRoutes(router, tweetHandlers)

	return router		
}

func setupUserRoutes(router *http.ServeMux, userHandlers *handlers.UserHandlers) {

}

func setupTweetRoutes(router *http.ServeMux, tweetHandlers *handlers.TweetHandlers) {

}