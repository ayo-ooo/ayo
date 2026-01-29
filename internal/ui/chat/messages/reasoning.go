package messages

import (
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// ReasoningCmp displays reasoning/thinking content with truncation and caching.
type ReasoningCmp struct {
	content       strings.Builder
	complete      bool
	dirty         bool
	rendered      string
	renderedWidth int
	maxLines      int
	mu            sync.RWMutex
}

// NewReasoningCmp creates a new reasoning component.
func NewReasoningCmp() *ReasoningCmp {
	return &ReasoningCmp{
		maxLines: 5, // Default to showing last 5 lines
	}
}

// Content returns the full reasoning content.
func (r *ReasoningCmp) Content() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.content.String()
}

// AppendContent adds more reasoning content.
func (r *ReasoningCmp) AppendContent(delta string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.content.WriteString(delta)
	r.dirty = true
	r.rendered = ""
}

// Reset clears the reasoning content.
func (r *ReasoningCmp) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.content.Reset()
	r.complete = false
	r.dirty = true
	r.rendered = ""
}

// SetComplete marks the reasoning as complete.
func (r *ReasoningCmp) SetComplete(complete bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.complete = complete
	r.dirty = true
	r.rendered = ""
}

// IsComplete returns whether reasoning is complete.
func (r *ReasoningCmp) IsComplete() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.complete
}

// IsEmpty returns whether there is any content.
func (r *ReasoningCmp) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.content.Len() == 0
}

// IsDirty returns whether the component needs re-rendering.
func (r *ReasoningCmp) IsDirty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.dirty
}

// Render returns the rendered reasoning display.
// Shows only the last maxLines lines for readability during streaming.
func (r *ReasoningCmp) Render(width int) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check cache
	if !r.dirty && r.rendered != "" && r.renderedWidth == width {
		return r.rendered
	}

	content := r.content.String()
	if content == "" {
		r.dirty = false
		return ""
	}

	// Truncate to last maxLines
	lines := strings.Split(content, "\n")
	if len(lines) > r.maxLines {
		lines = lines[len(lines)-r.maxLines:]
	}
	truncated := strings.Join(lines, "\n")

	// Style the output
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Italic(true)
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ca3af")).
		Italic(true).
		Width(width - 12). // Account for label
		MaxWidth(width - 12)

	r.rendered = labelStyle.Render("  Thinking: ") + contentStyle.Render(truncated)
	r.renderedWidth = width
	r.dirty = false

	return r.rendered
}
