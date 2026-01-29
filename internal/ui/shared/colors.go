// Package shared provides common UI components used by both interactive TUI
// and non-interactive streaming modes.
package shared

import "github.com/charmbracelet/lipgloss"

// Color palette - a cohesive dark theme inspired by popular terminal themes.
// These colors are used consistently across both TUI and non-interactive modes.
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#a78bfa") // Purple - main accent
	ColorSecondary = lipgloss.Color("#67e8f9") // Cyan - secondary accent
	ColorTertiary  = lipgloss.Color("#fbbf24") // Amber - warnings/tool labels
	ColorSuccess   = lipgloss.Color("#22c55e") // Green - success states (Tailwind green-500)
	ColorError     = lipgloss.Color("#ef4444") // Red - errors (Tailwind red-500)
	ColorMuted     = lipgloss.Color("#6b7280") // Gray - muted text
	ColorSubtle    = lipgloss.Color("#374151") // Dark gray - borders/backgrounds

	// Text colors
	ColorText       = lipgloss.Color("#e5e7eb") // Light gray - main text
	ColorTextDim    = lipgloss.Color("#9ca3af") // Medium gray - dim text
	ColorTextBright = lipgloss.Color("#f9fafb") // White - bright text

	// Background colors
	ColorBgDark   = lipgloss.Color("#1f2937") // Dark background
	ColorBgSubtle = lipgloss.Color("#111827") // Darker background
	ColorBgAccent = lipgloss.Color("#312e81") // Purple tinted background

	// Tool-specific colors
	ColorToolName    = lipgloss.Color("#3b82f6") // Blue - tool names
	ColorToolPending = lipgloss.Color("#6b7280") // Gray - pending state
	ColorToolRunning = lipgloss.Color("#a78bfa") // Purple - running state
)
