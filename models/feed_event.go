package models

import "time"

// FeedEventType represents the type of event for the feed SSE
// e.g. tweet created, liked, retweeted
type FeedEventType string

const (
	FeedEventCreated   FeedEventType = "created"
	FeedEventLiked     FeedEventType = "liked"
	FeedEventRetweeted FeedEventType = "retweeted"
)

// FeedEvent represents an event in the user's feed for SSE
type FeedEvent struct {
	Type      FeedEventType `json:"type"`
	Tweet     TweetProps    `json:"tweet"`
	ActorID   string        `json:"actor_id"` // The user who performed the action
	CreatedAt time.Time     `json:"created_at"`
}
