package notifications

import (
	"log"
	"sync"

	"github.com/aimrintech/x-backend/models"
)

type Notifications interface {
	Subscribe(topic models.NotificationType, userID string) <-chan models.Notification
	Unsubscribe(topic models.NotificationType, userID string)
	Publish(topic models.NotificationType, notification *models.Notification)
}

type NotificationsService struct {
	subscribers map[models.NotificationType]map[string]chan models.Notification
	mu          sync.RWMutex
}

func NewNotificationsService() Notifications {
	return &NotificationsService{
		subscribers: make(map[models.NotificationType]map[string]chan models.Notification),
	}
}

func (s *NotificationsService) Subscribe(topic models.NotificationType, userID string) <-chan models.Notification {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscribers[topic]; !ok {
		s.subscribers[topic] = make(map[string]chan models.Notification)
	}

	ch := make(chan models.Notification)
	s.subscribers[topic][userID] = ch

	log.Printf("Subscribed to %s for user %s", topic, userID)

	return ch
}

func (s *NotificationsService) Unsubscribe(topic models.NotificationType, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userChans, ok := s.subscribers[topic]
	if !ok {
		return
	}

	ch, ok := userChans[userID]
	if !ok {
		return
	}

	close(ch)
	delete(userChans, userID)
	if len(userChans) == 0 {
		delete(s.subscribers, topic)
	}

	log.Printf("Unsubscribed from %s for user %s", topic, userID)
}

func (s *NotificationsService) Publish(topic models.NotificationType, notification *models.Notification) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ch, ok := s.subscribers[topic][notification.TargetUserID]
	if !ok {
		return
	}

	select {
	case ch <- *notification:
	default:
		log.Printf("Dropping notification for user %s: channel full", notification.TargetUserID)
	}
}
