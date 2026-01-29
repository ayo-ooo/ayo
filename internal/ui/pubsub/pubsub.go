// Package pubsub provides a generic event broker for decoupled communication
// between TUI components. It follows the pattern from Crush for streaming
// LLM responses, tool calls, and other events to the UI.
package pubsub

import (
	"context"
	"sync"
)

// EventType represents the kind of event being published.
type EventType string

const (
	// CreatedEvent indicates a new entity was created.
	CreatedEvent EventType = "created"
	// UpdatedEvent indicates an existing entity was updated.
	UpdatedEvent EventType = "updated"
	// DeletedEvent indicates an entity was deleted.
	DeletedEvent EventType = "deleted"
	// StartedEvent indicates an operation has started.
	StartedEvent EventType = "started"
	// CompletedEvent indicates an operation completed successfully.
	CompletedEvent EventType = "completed"
	// FailedEvent indicates an operation failed.
	FailedEvent EventType = "failed"
	// CancelledEvent indicates an operation was cancelled.
	CancelledEvent EventType = "cancelled"
)

// Event represents a typed event with a payload.
type Event[T any] struct {
	Type    EventType
	Payload T
}

// Broker manages subscriptions and publishes events to subscribers.
// It is generic over the payload type T.
type Broker[T any] struct {
	mu          sync.RWMutex
	subscribers map[chan Event[T]]struct{}
	bufferSize  int
}

// NewBroker creates a new broker with the specified channel buffer size.
// A larger buffer allows more events to queue before blocking.
func NewBroker[T any](bufferSize int) *Broker[T] {
	if bufferSize < 1 {
		bufferSize = 16
	}
	return &Broker[T]{
		subscribers: make(map[chan Event[T]]struct{}),
		bufferSize:  bufferSize,
	}
}

// Subscribe creates a new subscription channel. The channel will receive
// events until the context is cancelled, at which point it is automatically
// unsubscribed and closed.
func (b *Broker[T]) Subscribe(ctx context.Context) <-chan Event[T] {
	ch := make(chan Event[T], b.bufferSize)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	// Unsubscribe when context is cancelled
	go func() {
		<-ctx.Done()
		b.unsubscribe(ch)
	}()

	return ch
}

// unsubscribe removes a channel from the subscriber set and closes it.
func (b *Broker[T]) unsubscribe(ch chan Event[T]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Publish sends an event to all subscribers. It uses non-blocking sends
// to prevent slow subscribers from blocking the publisher. Events are
// dropped if a subscriber's buffer is full.
func (b *Broker[T]) Publish(event Event[T]) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- event:
			// Event sent successfully
		default:
			// Subscriber buffer full, drop event
		}
	}
}

// PublishSync sends an event to all subscribers and blocks until all
// have received it. Use sparingly as slow subscribers will block.
func (b *Broker[T]) PublishSync(event Event[T]) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		ch <- event
	}
}

// SubscriberCount returns the current number of subscribers.
func (b *Broker[T]) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
