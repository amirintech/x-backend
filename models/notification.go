package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeLike    NotificationType = "like"
	NotificationTypeFollow  NotificationType = "follow"
	NotificationTypeRetweet NotificationType = "retweet"
	NotificationTypeReply   NotificationType = "reply"
	NotificationTypeMention NotificationType = "mention"
)

type Notification struct {
	ID            string           `json:"id"`
	TargetUserID  string           `json:"target_user_id"`
	TargetTweetID *string          `json:"target_tweet_id"`
	AuthorUserID  string           `json:"author_user_id"`
	Type          NotificationType `json:"type"`
	IsRead        bool             `json:"is_read"`
	CreatedAt     time.Time        `json:"created_at"`
}

func NewNotification(targetUserID, authorUserID string, targetTweetID *string, notificationType NotificationType) *Notification {
	return &Notification{
		ID:            uuid.New().String(),
		TargetUserID:  targetUserID,
		TargetTweetID: targetTweetID,
		AuthorUserID:  authorUserID,
		Type:          notificationType,
		IsRead:        false,
		CreatedAt:     time.Now(),
	}
}
