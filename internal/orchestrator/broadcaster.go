package orchestrator

import (
	"sync"

	"github.com/rs/zerolog/log"
)

// Broadcaster handles broadcasting messages to subscribers
type Broadcaster struct {
	orchestrator *Orchestrator
	subscribers  map[string]chan BroadcastMessage
	mu           sync.RWMutex
}

// NewBroadcaster creates a new broadcaster
func NewBroadcaster(o *Orchestrator) *Broadcaster {
	return &Broadcaster{
		orchestrator: o,
		subscribers:  make(map[string]chan BroadcastMessage),
	}
}

// Subscribe adds a new subscriber and returns a channel for receiving messages
func (b *Broadcaster) Subscribe(id string) chan BroadcastMessage {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Create buffered channel for this subscriber
	// Increased buffer size to handle high-frequency price updates
	ch := make(chan BroadcastMessage, 1000)
	b.subscribers[id] = ch

	log.Debug().Str("subscriberID", id).Int("totalSubscribers", len(b.subscribers)).Msg("Subscriber added")

	return ch
}

// Unsubscribe removes a subscriber
func (b *Broadcaster) Unsubscribe(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, exists := b.subscribers[id]; exists {
		close(ch)
		delete(b.subscribers, id)
		log.Debug().Str("subscriberID", id).Int("totalSubscribers", len(b.subscribers)).Msg("Subscriber removed")
	}
}

// Broadcast sends a message to all subscribers
func (b *Broadcaster) Broadcast(msg BroadcastMessage) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for id, ch := range b.subscribers {
		select {
		case ch <- msg:
			// Message sent successfully
		default:
			// Channel is full, skip this message to avoid blocking
			log.Warn().Str("subscriberID", id).Str("messageType", msg.Type).Msg("Subscriber channel full, message dropped")
		}
	}
}

// GetSubscriberCount returns the number of active subscribers
func (b *Broadcaster) GetSubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}

// Close closes all subscriber channels
func (b *Broadcaster) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for id, ch := range b.subscribers {
		close(ch)
		delete(b.subscribers, id)
	}

	log.Info().Msg("Broadcaster closed, all subscribers removed")
}
