package messages

import (
	"strings"
	"sync"
)

// MessageList manages a collection of message components with efficient rendering.
type MessageList struct {
	messages []MessageComponent
	mu       sync.RWMutex
}

// NewMessageList creates a new message list.
func NewMessageList() *MessageList {
	return &MessageList{
		messages: make([]MessageComponent, 0, 32),
	}
}

// Add appends a message to the list.
func (l *MessageList) Add(msg MessageComponent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, msg)
}

// GetByID returns the message with the given ID, or nil if not found.
func (l *MessageList) GetByID(id string) MessageComponent {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, msg := range l.messages {
		if msg.ID() == id {
			return msg
		}
	}
	return nil
}

// GetLast returns the last message, or nil if empty.
func (l *MessageList) GetLast() MessageComponent {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.messages) == 0 {
		return nil
	}
	return l.messages[len(l.messages)-1]
}

// Count returns the number of messages.
func (l *MessageList) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.messages)
}

// Render returns the combined rendered output of all messages.
// Each message uses its cache, so this is efficient even with many messages.
func (l *MessageList) Render(width int) string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.messages) == 0 {
		return ""
	}

	var b strings.Builder
	for i, msg := range l.messages {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(msg.Render(width))
	}
	return b.String()
}

// Clear removes all messages from the list.
func (l *MessageList) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = l.messages[:0]
}

// All returns a copy of all messages (for iteration).
func (l *MessageList) All() []MessageComponent {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]MessageComponent, len(l.messages))
	copy(result, l.messages)
	return result
}
