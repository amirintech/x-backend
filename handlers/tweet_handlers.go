package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/stores"
)

type TweetHandlers struct {
	tweetStore *stores.TweetStore
}

func NewTweetHandlers(tweetStore *stores.TweetStore) *TweetHandlers {
	return &TweetHandlers{
		tweetStore: tweetStore,
	}
}

func (h *TweetHandlers) GetTweets(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	limit, err := strconv.Atoi(params.Get("limit"))
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(params.Get("offset"))
	if err != nil {
		offset = 0
	}

	tweets, err := (*h.tweetStore).GetTweets(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get tweets")
		return
	}

	writeJSON(w, http.StatusOK, tweets)
}

func (h *TweetHandlers) GetTweetByID(w http.ResponseWriter, r *http.Request) {
	tweetID := r.PathValue("id")
	if tweetID == "" {
		writeError(w, http.StatusBadRequest, "Tweet ID is required")
		return
	}

	tweet, err := (*h.tweetStore).GetTweetByID(tweetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get tweet")
		return
	}

	writeJSON(w, http.StatusOK, tweet)
}

func (h *TweetHandlers) CreateTweet(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	fmt.Println("userID", userID)

	type CreateTweetRequest struct {
		Content   *string   `json:"content"`
		MediaURLs *[]string `json:"media_urls"`
		Hashtags  *[]string `json:"hashtags"`
	}

	var req CreateTweetRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	fmt.Println("req", req)

	if req.Content == nil && req.MediaURLs == nil {
		writeError(w, http.StatusBadRequest, "Content or media URLs are required")
		return
	}

	if req.Content != nil && len(*req.Content) > 280 {
		writeError(w, http.StatusBadRequest, "Content must be less than 280 characters")
		return
	}

	t := models.Tweet{
		Content:   req.Content,
		MediaURLs: req.MediaURLs,
		Hashtags:  req.Hashtags,
	}
	tweet, err := (*h.tweetStore).CreateTweet(&t, userID)
	if err != nil {
		fmt.Println("err", err)
		writeError(w, http.StatusInternalServerError, "Failed to create tweet")
		return
	}

	fmt.Println("tweet", tweet)

	writeJSON(w, http.StatusOK, tweet)
}

func (h *TweetHandlers) LikeTweet(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tweetID := r.PathValue("id")
	if tweetID == "" {
		writeError(w, http.StatusBadRequest, "Tweet ID is required")
		return
	}

	err = (*h.tweetStore).LikeTweet(tweetID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to like tweet")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Tweet liked"})
}

func (h *TweetHandlers) UnlikeTweet(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tweetID := r.PathValue("id")
	if tweetID == "" {
		writeError(w, http.StatusBadRequest, "Tweet ID is required")
		return
	}

	err = (*h.tweetStore).UnlikeTweet(tweetID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to unlike tweet")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Tweet unliked"})
}
