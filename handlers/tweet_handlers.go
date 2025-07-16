package handlers

import (
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
	limit, offset := extractPaginationParams(r)

	tweets, err := (*h.tweetStore).GetTweets(limit, offset)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to get tweets")
		return
	}

	writeJSON(w, r, http.StatusOK, tweets)
}

func (h *TweetHandlers) GetUsersWithTweets(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit, offset := extractPaginationParams(r)

	usersWithTweets, err := (*h.tweetStore).GetUsersWithTweets(userID, limit, offset)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to get users with tweets")
		return
	}

	writeJSON(w, r, http.StatusOK, usersWithTweets)
}

func (h *TweetHandlers) GetTweetByID(w http.ResponseWriter, r *http.Request) {
	tweetID := r.PathValue("id")
	if tweetID == "" {
		writeError(w, r, http.StatusBadRequest, "Tweet ID is required")
		return
	}

	tweet, err := (*h.tweetStore).GetTweetByID(tweetID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to get tweet")
		return
	}

	writeJSON(w, r, http.StatusOK, tweet)
}

func (h *TweetHandlers) CreateTweet(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type CreateTweetRequest struct {
		Content   *string   `json:"content"`
		MediaURLs *[]string `json:"mediaURLs"`
	}

	var req CreateTweetRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == nil && req.MediaURLs == nil {
		writeError(w, r, http.StatusBadRequest, "Content or media URLs are required")
		return
	}

	if req.Content != nil && len(*req.Content) > 280 {
		writeError(w, r, http.StatusBadRequest, "Content must be less than 280 characters")
		return
	}

	t := models.Tweet{
		Content:   req.Content,
		MediaURLs: req.MediaURLs,
	}
	tweet, err := (*h.tweetStore).CreateTweet(&t, userID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to create tweet")
		return
	}

	writeJSON(w, r, http.StatusOK, tweet)
}

func (h *TweetHandlers) LikeTweet(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tweetID := r.PathValue("id")
	if tweetID == "" {
		writeError(w, r, http.StatusBadRequest, "Tweet ID is required")
		return
	}

	err = (*h.tweetStore).LikeTweet(tweetID, userID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to like tweet")
		return
	}

	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Tweet liked"})
}

func (h *TweetHandlers) UnlikeTweet(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tweetID := r.PathValue("id")
	if tweetID == "" {
		writeError(w, r, http.StatusBadRequest, "Tweet ID is required")
		return
	}

	err = (*h.tweetStore).UnlikeTweet(tweetID, userID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to unlike tweet")
		return
	}

	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Tweet unliked"})
}

func extractPaginationParams(r *http.Request) (int, int) {
	params := r.URL.Query()
	limit, err := strconv.Atoi(params.Get("limit"))
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(params.Get("offset"))
	if err != nil {
		offset = 0
	}
	return limit, offset
}
