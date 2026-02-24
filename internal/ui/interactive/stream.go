// Package interactive provides a simplified streaming text renderer
// for interactive mode that writes tokens directly to the terminal.
package interactive

import (
	"io"
	"strings"
	"sync"

	"github.com/alexcabrera/ayo/internal/ui/shared"
	"github.com/charmbracelet/glamour"
)

// StreamRenderer handles streaming text output with markdown rendering.
// It buffers partial lines and renders complete lines with glamour.
type StreamRenderer struct {
	out     io.Writer
	glamour *glamour.TermRenderer
	buffer  strings.Builder
	width   int
	mu      sync.Mutex
}

// NewStreamRenderer creates a new streaming renderer.
// Width controls the markdown rendering width (clamped to 40-200).
func NewStreamRenderer(out io.Writer, width int) *StreamRenderer {
	return &StreamRenderer{
		out:     out,
		glamour: shared.GetStyledMarkdownRenderer(width),
		width:   shared.Clamp(width, 40, 200),
	}
}

// WriteToken appends a token to the buffer and renders complete lines.
// Partial lines are buffered until a newline is received.
func (r *StreamRenderer) WriteToken(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buffer.WriteString(token)

	// Check for complete lines
	content := r.buffer.String()
	if idx := strings.LastIndex(content, "\n"); idx >= 0 {
		complete := content[:idx+1]
		r.buffer.Reset()
		r.buffer.WriteString(content[idx+1:])

		// Render and write complete lines
		if r.glamour != nil {
			rendered, err := r.glamour.Render(complete)
			if err != nil {
				// Fall back to raw output on render error
				_, err = r.out.Write([]byte(complete))
				return err
			}
			_, err = r.out.Write([]byte(rendered))
			return err
		}
		// No glamour renderer, write raw
		_, err := r.out.Write([]byte(complete))
		return err
	}

	return nil
}

// Flush renders any remaining buffered content.
// Call this when streaming is complete.
func (r *StreamRenderer) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.buffer.Len() == 0 {
		return nil
	}

	content := r.buffer.String()
	r.buffer.Reset()

	if r.glamour != nil {
		rendered, err := r.glamour.Render(content)
		if err != nil {
			// Fall back to raw output on render error
			_, err = r.out.Write([]byte(content))
			return err
		}
		_, err = r.out.Write([]byte(rendered))
		return err
	}

	_, err := r.out.Write([]byte(content))
	return err
}

// Reset clears the buffer without flushing.
func (r *StreamRenderer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buffer.Reset()
}

// Buffered returns the current buffered content (for testing).
func (r *StreamRenderer) Buffered() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buffer.String()
}
