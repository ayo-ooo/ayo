// Package styles provides theming and styling for TUI components.
package styles

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color palette and styles for the application.
type Theme struct {
	// Primary colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color

	// Status colors
	Green      lipgloss.Color
	GreenDark  lipgloss.Color
	GreenLight lipgloss.Color
	Red        lipgloss.Color
	RedDark    lipgloss.Color
	Blue       lipgloss.Color
	BlueDark   lipgloss.Color
	BlueLight  lipgloss.Color
	Yellow     lipgloss.Color
	Info       lipgloss.Color

	// Foreground colors
	FgBase      lipgloss.Color
	FgMuted     lipgloss.Color
	FgHalfMuted lipgloss.Color
	FgSubtle    lipgloss.Color

	// Background colors
	BgBase        lipgloss.Color
	BgBaseLighter lipgloss.Color
	Border        lipgloss.Color
	White         lipgloss.Color

	// Cached styles
	styles *Styles
}

// Styles holds pre-built styles for common use cases.
type Styles struct {
	Base   lipgloss.Style
	Muted  lipgloss.Style
	Subtle lipgloss.Style
}

// S returns the cached styles accessor.
func (t *Theme) S() *Styles {
	if t.styles == nil {
		t.styles = &Styles{
			Base:   lipgloss.NewStyle().Foreground(t.FgBase),
			Muted:  lipgloss.NewStyle().Foreground(t.FgMuted),
			Subtle: lipgloss.NewStyle().Foreground(t.FgSubtle),
		}
	}
	return t.styles
}

// FieldStyles defines styles for focused/blurred states.
type FieldStyles struct {
	Base        lipgloss.Style
	Title       lipgloss.Style
	Description lipgloss.Style
	Border      lipgloss.Style
	Cursor      lipgloss.Style
}

// FocusedStyles returns styles for focused components.
func (t *Theme) FocusedStyles() FieldStyles {
	return FieldStyles{
		Base:        lipgloss.NewStyle().Foreground(t.FgBase),
		Title:       lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Description: lipgloss.NewStyle().Foreground(t.FgMuted),
		Border:      lipgloss.NewStyle().Foreground(t.Primary),
		Cursor:      lipgloss.NewStyle().Foreground(t.Primary),
	}
}

// BlurredStyles returns styles for unfocused components.
func (t *Theme) BlurredStyles() FieldStyles {
	return FieldStyles{
		Base:        lipgloss.NewStyle().Foreground(t.FgMuted),
		Title:       lipgloss.NewStyle().Foreground(t.FgMuted),
		Description: lipgloss.NewStyle().Foreground(t.FgSubtle),
		Border:      lipgloss.NewStyle().Foreground(t.Border),
		Cursor:      lipgloss.NewStyle().Foreground(t.FgMuted),
	}
}

// Status icons.
const (
	ToolPending = "○"
	ToolSuccess = "●"
	ToolError   = "×"
	ToolRunning = "◐"
	ArrowRight  = "▸"
	CheckMark   = "✓"
)

// DefaultTheme returns the default dark theme.
func DefaultTheme() *Theme {
	return &Theme{
		Primary:   lipgloss.Color("#a78bfa"),
		Secondary: lipgloss.Color("#67e8f9"),

		Green:      lipgloss.Color("#22c55e"),
		GreenDark:  lipgloss.Color("#16a34a"),
		GreenLight: lipgloss.Color("#86efac"),
		Red:        lipgloss.Color("#ef4444"),
		RedDark:    lipgloss.Color("#dc2626"),
		Blue:       lipgloss.Color("#3b82f6"),
		BlueDark:   lipgloss.Color("#2563eb"),
		BlueLight:  lipgloss.Color("#93c5fd"),
		Yellow:     lipgloss.Color("#eab308"),
		Info:       lipgloss.Color("#0ea5e9"),

		FgBase:      lipgloss.Color("#f4f4f5"),
		FgMuted:     lipgloss.Color("#a1a1aa"),
		FgHalfMuted: lipgloss.Color("#71717a"),
		FgSubtle:    lipgloss.Color("#52525b"),

		BgBase:        lipgloss.Color("#18181b"),
		BgBaseLighter: lipgloss.Color("#27272a"),
		Border:        lipgloss.Color("#3f3f46"),
		White:         lipgloss.Color("#ffffff"),
	}
}

// currentTheme is the active theme.
var currentTheme = DefaultTheme()

// CurrentTheme returns the current theme.
func CurrentTheme() *Theme {
	return currentTheme
}

// SetTheme sets the current theme.
func SetTheme(t *Theme) {
	currentTheme = t
}

// markdownRenderer caches the glamour renderer.
var markdownRenderer *glamour.TermRenderer

// GetPlainMarkdownRenderer returns a markdown renderer for the given width.
func GetPlainMarkdownRenderer(width int) *glamour.TermRenderer {
	if markdownRenderer == nil || width > 0 {
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
		)
		if err != nil {
			return nil
		}
		markdownRenderer = r
	}
	return markdownRenderer
}

// RenderMarkdown renders markdown content.
func RenderMarkdown(content string, width int) string {
	r := GetPlainMarkdownRenderer(width)
	if r == nil {
		return content
	}
	rendered, err := r.Render(content)
	if err != nil {
		return content
	}
	return rendered
}
