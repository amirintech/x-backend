package handlers

import "github.com/aimrintech/x-backend/stores"

type TweetHandlers struct {
	tweetStore *stores.TweetStore
}

func NewTweetHandlers(tweetStore *stores.TweetStore) *TweetHandlers {
	return &TweetHandlers{
		tweetStore: tweetStore,
	}
}
