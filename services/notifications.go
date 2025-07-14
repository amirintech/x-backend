package services

import (
	"github.com/aimrintech/x-backend/models"
)

type Notifications interface {
	Subscribe(topic models.NotificationType, userID string) <-chan models.Notification
	Unsubscribe(topic models.NotificationType, userID string)
	Publish(topic models.NotificationType, notification *models.Notification)
}

type NotificationsService struct {
	sse *SSEService
}

func NewNotificationsService() Notifications {
	return &NotificationsService{
		sse: NewSSEService(),
	}
}

func (s *NotificationsService) Subscribe(topic models.NotificationType, userID string) <-chan models.Notification {
	ch := s.sse.Subscribe(string(topic), userID)
	out := make(chan models.Notification)
	go func() {
		for msg := range ch {
			if notif, ok := msg.(models.Notification); ok {
				out <- notif
			}
		}
		close(out)
	}()
	return out
}

func (s *NotificationsService) Unsubscribe(topic models.NotificationType, userID string) {
	s.sse.Unsubscribe(string(topic), userID)
}

func (s *NotificationsService) Publish(topic models.NotificationType, notification *models.Notification) {
	s.sse.Publish(string(topic), notification.TargetUserID, *notification)
}
