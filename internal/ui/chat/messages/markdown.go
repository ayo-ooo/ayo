package messages

import (
	"sync"

	"github.com/charmbracelet/glamour"
)

// rendererCache stores glamour renderers by width to avoid recreating them.
var rendererCache sync.Map // map[int]*glamour.TermRenderer

// GetMarkdownRenderer returns a cached glamour renderer for the given width.
// Width is clamped to 40-200 to prevent cache explosion.
func GetMarkdownRenderer(width int) *glamour.TermRenderer {
	// Clamp width to reasonable range
	width = clamp(width, 40, 200)

	// Check cache
	if r, ok := rendererCache.Load(width); ok {
		return r.(*glamour.TermRenderer)
	}

	// Create new renderer
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}

	// Store and return (may race, but that's fine - both valid)
	rendererCache.Store(width, r)
	return r
}

// clamp constrains a value to a range.
func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// ClearRendererCache clears the renderer cache (useful for testing).
func ClearRendererCache() {
	rendererCache.Range(func(key, value any) bool {
		rendererCache.Delete(key)
		return true
	})
}
