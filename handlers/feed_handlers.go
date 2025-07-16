package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aimrintech/x-backend/services"
)

type FeedHandlers struct {
	feedService services.Feed
}

func NewFeedHandlers(feedService services.Feed) *FeedHandlers {
	return &FeedHandlers{
		feedService: feedService,
	}
}

func (h *FeedHandlers) StreamFeed(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("StreamFeed handler called: %s %s\n", r.Method, r.URL.Path)
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if userID == "" {
		writeError(w, r, http.StatusBadRequest, "User ID is required")
		return
	}

	fmt.Printf("StreamFeed: Starting SSE for user %s\n", userID)
	setStreamHeaders(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Printf("StreamFeed: Flusher not supported\n")
		writeError(w, r, http.StatusInternalServerError, "Streaming unsupported")
		return
	}

	feedChan := h.feedService.Subscribe(userID)
	defer h.feedService.Unsubscribe(userID)

	ctx := r.Context()
	fmt.Printf("StreamFeed: Entering SSE loop for user %s\n", userID)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("StreamFeed: Context done for user %s\n", userID)
			return
		case event, ok := <-feedChan:
			if !ok {
				fmt.Printf("StreamFeed: Channel closed for user %s\n", userID)
				return
			}
			fmt.Printf("StreamFeed: Received event for user %s: %+v\n", userID, event)
			data, err := json.Marshal(event)
			if err != nil {
				fmt.Printf("StreamFeed: JSON marshal error: %v\n", err)
				continue
			}
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

func setStreamHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
