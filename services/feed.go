package services

import (
	"sync"

	"github.com/aimrintech/x-backend/models"
)

type Feed interface {
	Subscribe(userID string) <-chan models.FeedEvent
	Unsubscribe(userID string)
	Publish(userID string, event *models.FeedEvent)
	PublishToAll(event *models.FeedEvent)
	Close()
}

type FeedService struct {
	subscribers map[string][]chan models.FeedEvent
	mu          sync.RWMutex
	closed      bool
}

func NewFeedService() Feed {
	return &FeedService{
		subscribers: make(map[string][]chan models.FeedEvent),
	}
}

func (s *FeedService) Subscribe(userID string) <-chan models.FeedEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		ch := make(chan models.FeedEvent)
		close(ch)
		return ch
	}

	ch := make(chan models.FeedEvent, 100) // Buffered channel to prevent blocking
	s.subscribers[userID] = append(s.subscribers[userID], ch)
	return ch
}

func (s *FeedService) Unsubscribe(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if channels, exists := s.subscribers[userID]; exists {
		// Close all channels for this user
		for _, ch := range channels {
			close(ch)
		}
		delete(s.subscribers, userID)
	}
}

func (s *FeedService) Publish(userID string, event *models.FeedEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return
	}

	if channels, exists := s.subscribers[userID]; exists {
		// Send to all channels for this specific user
		for i := len(channels) - 1; i >= 0; i-- {
			select {
			case channels[i] <- *event:
				// Successfully sent
			default:
				// Channel is full or closed, remove it
				s.removeChannel(userID, i)
			}
		}
	}
}

func (s *FeedService) PublishToAll(event *models.FeedEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return
	}

	// Send to all subscribers
	for userID, channels := range s.subscribers {
		for i := len(channels) - 1; i >= 0; i-- {
			select {
			case channels[i] <- *event:
				// Successfully sent
			default:
				// Channel is full or closed, remove it
				s.removeChannel(userID, i)
			}
		}
	}
}

func (s *FeedService) removeChannel(userID string, index int) {
	// This should be called with write lock already held
	if channels, exists := s.subscribers[userID]; exists && index < len(channels) {
		close(channels[index])
		// Remove the channel from the slice
		s.subscribers[userID] = append(channels[:index], channels[index+1:]...)
		// If no channels left for this user, remove the user entry
		if len(s.subscribers[userID]) == 0 {
			delete(s.subscribers, userID)
		}
	}
}

func (s *FeedService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true

	// Close all channels
	for userID, channels := range s.subscribers {
		for _, ch := range channels {
			close(ch)
		}
		delete(s.subscribers, userID)
	}
}
