// Package messages provides components for rendering chat messages,
// tool calls, and other content in the TUI.
package messages

import (
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// MessageComponent defines the interface for message display components.
// Each message caches its rendered output and only re-renders when dirty
// or when the width changes.
type MessageComponent interface {
	// ID returns the unique identifier for this message.
	ID() string

	// Role returns the message role ("user" or "assistant").
	Role() string

	// Content returns the raw message content.
	Content() string

	// SetContent replaces the message content entirely.
	SetContent(content string)

	// AppendContent appends to the message content (for streaming).
	AppendContent(delta string)

	// Render returns the rendered message for the given width.
	// Uses cached output if available and not dirty.
	Render(width int) string

	// IsComplete returns whether the message is fully received.
	IsComplete() bool

	// SetComplete marks the message as complete.
	SetComplete(complete bool)

	// IsDirty returns whether the message needs re-rendering.
	IsDirty() bool
}

// baseMessageCmp provides common functionality for all message components.
// It handles caching and dirty tracking.
type baseMessageCmp struct {
	id            string
	role          string
	content       strings.Builder
	complete      bool
	dirty         bool
	rendered      string
	renderedWidth int
	mu            sync.RWMutex
}

// ID returns the message ID.
func (m *baseMessageCmp) ID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.id
}

// Role returns the message role.
func (m *baseMessageCmp) Role() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.role
}

// Content returns the raw message content.
func (m *baseMessageCmp) Content() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.content.String()
}

// SetContent replaces the message content entirely.
func (m *baseMessageCmp) SetContent(content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.content.Reset()
	m.content.WriteString(content)
	m.dirty = true
	m.rendered = ""
}

// AppendContent appends to the message content (for streaming).
func (m *baseMessageCmp) AppendContent(delta string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.content.WriteString(delta)
	m.dirty = true
	m.rendered = ""
}

// IsComplete returns whether the message is fully received.
func (m *baseMessageCmp) IsComplete() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.complete
}

// SetComplete marks the message as complete.
func (m *baseMessageCmp) SetComplete(complete bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.complete = complete
}

// IsDirty returns whether the message needs re-rendering.
func (m *baseMessageCmp) IsDirty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dirty
}

// checkCache returns the cached render if valid, otherwise empty string.
// Caller must hold at least read lock.
func (m *baseMessageCmp) checkCache(width int) (string, bool) {
	if !m.dirty && m.rendered != "" && m.renderedWidth == width {
		return m.rendered, true
	}
	return "", false
}

// updateCache stores the rendered output.
// Caller must hold write lock.
func (m *baseMessageCmp) updateCache(rendered string, width int) {
	m.rendered = rendered
	m.renderedWidth = width
	m.dirty = false
}

// UserMessageCmp renders user messages with a simple prefix style.
type UserMessageCmp struct {
	baseMessageCmp
}

// NewUserMessage creates a new user message component.
func NewUserMessage(id string) *UserMessageCmp {
	return &UserMessageCmp{
		baseMessageCmp: baseMessageCmp{
			id:       id,
			role:     "user",
			complete: true, // User messages are always complete
		},
	}
}

// Render returns the rendered user message.
func (m *UserMessageCmp) Render(width int) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cached, ok := m.checkCache(width); ok {
		return cached
	}

	// Simple styled output for user messages
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7dd3fc")).
		Bold(true)

	label := style.Render("You")
	content := m.content.String()

	// Wrap content if needed
	contentStyle := lipgloss.NewStyle().
		Width(width - 2)
	wrapped := contentStyle.Render(content)

	rendered := label + "\n" + wrapped
	m.updateCache(rendered, width)

	return rendered
}

// AssistantMessageCmp renders assistant messages with markdown styling.
type AssistantMessageCmp struct {
	baseMessageCmp
	agentHandle string
}

// NewAssistantMessage creates a new assistant message component.
func NewAssistantMessage(id, agentHandle string) *AssistantMessageCmp {
	return &AssistantMessageCmp{
		baseMessageCmp: baseMessageCmp{
			id:   id,
			role: "assistant",
		},
		agentHandle: agentHandle,
	}
}

// AgentHandle returns the agent handle for this message.
func (m *AssistantMessageCmp) AgentHandle() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.agentHandle
}

// Render returns the rendered assistant message with markdown processing.
func (m *AssistantMessageCmp) Render(width int) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cached, ok := m.checkCache(width); ok {
		return cached
	}

	// Render label
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a78bfa")).
		Bold(true)
	label := labelStyle.Render(m.agentHandle)

	// Render content with markdown
	content := m.content.String()
	if content == "" {
		// Empty streaming message - show just the label
		m.updateCache(label, width)
		return label
	}

	// Use cached glamour renderer
	contentWidth := width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	renderer := GetMarkdownRenderer(contentWidth)
	var rendered string
	if renderer != nil {
		md, err := renderer.Render(content)
		if err == nil {
			rendered = strings.TrimSpace(md)
		} else {
			rendered = content
		}
	} else {
		rendered = content
	}

	result := label + "\n" + rendered
	m.updateCache(result, width)

	return result
}
