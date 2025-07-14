package services

import (
	"log"
	"sync"
)

type SSE interface {
	Subscribe(topic string, userID string) <-chan interface{}
	Unsubscribe(topic string, userID string)
	Publish(topic string, userID string, message interface{})
}

type SSEService struct {
	subscribers map[string]map[string]chan interface{}
	mu          sync.RWMutex
}

func NewSSEService() *SSEService {
	return &SSEService{
		subscribers: make(map[string]map[string]chan interface{}),
	}
}

func (s *SSEService) Subscribe(topic string, userID string) <-chan interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscribers[topic]; !ok {
		s.subscribers[topic] = make(map[string]chan interface{})
	}

	ch := make(chan interface{})
	s.subscribers[topic][userID] = ch

	log.Printf("Subscribed to %s for user %s", topic, userID)

	return ch
}

func (s *SSEService) Unsubscribe(topic string, userID string) {
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

func (s *SSEService) Publish(topic string, userID string, message interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ch, ok := s.subscribers[topic][userID]
	if !ok {
		return
	}

	select {
	case ch <- message:
	default:
		log.Printf("Dropping message for user %s on topic %s: channel full", userID, topic)
	}

	log.Printf("Published message to %s for user %s", topic, userID)
}
