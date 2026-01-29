package ui

import (
	"os"
	"strings"

	"github.com/alexcabrera/ayo/internal/ui/shared"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Color palette - re-exported from shared package for backward compatibility.
var (
	// Primary colors
	colorPrimary   = shared.ColorPrimary
	colorSecondary = shared.ColorSecondary
	colorTertiary  = shared.ColorTertiary
	colorSuccess   = shared.ColorSuccess
	colorError     = shared.ColorError
	colorMuted     = shared.ColorMuted
	colorSubtle    = shared.ColorSubtle

	// Text colors
	colorText       = shared.ColorText
	colorTextDim    = shared.ColorTextDim
	colorTextBright = shared.ColorTextBright

	// Background colors
	colorBgDark   = shared.ColorBgDark
	colorBgSubtle = shared.ColorBgSubtle
	colorBgAccent = shared.ColorBgAccent
)

// Styles holds all the application styles.
type Styles struct {
	// Section labels
	ReasoningLabel lipgloss.Style
	ToolLabel      lipgloss.Style
	ErrorLabel     lipgloss.Style
	SuccessLabel   lipgloss.Style
	InfoLabel      lipgloss.Style

	// Content boxes
	ReasoningBox lipgloss.Style
	ToolBox      lipgloss.Style
	ErrorBox     lipgloss.Style
	CodeBox      lipgloss.Style

	// Text styles
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Command    lipgloss.Style
	FilePath   lipgloss.Style
	Muted      lipgloss.Style
	Emphasis   lipgloss.Style
	Bold       lipgloss.Style

	// Status indicators
	StatusPending    lipgloss.Style
	StatusInProgress lipgloss.Style
	StatusComplete   lipgloss.Style
	StatusError      lipgloss.Style

	// Borders
	BorderActive lipgloss.Border
	BorderNormal lipgloss.Border

	// Width constraint
	MaxWidth int
}

// DefaultStyles returns the default application styles.
func DefaultStyles() Styles {
	maxWidth := getTerminalWidth()
	if maxWidth > 120 {
		maxWidth = 120
	}

	return Styles{
		// Section labels with icons
		ReasoningLabel: lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			MarginBottom(1),

		ToolLabel: lipgloss.NewStyle().
			Foreground(colorTertiary).
			Bold(true).
			MarginBottom(1),

		ErrorLabel: lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true),

		SuccessLabel: lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true),

		InfoLabel: lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true),

		// Content boxes
		ReasoningBox: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorSubtle).
			Padding(1, 2).
			MarginBottom(1).
			MaxWidth(maxWidth),

		ToolBox: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorSubtle).
			BorderLeft(true).
			BorderRight(false).
			BorderTop(false).
			BorderBottom(false).
			PaddingLeft(2).
			MarginLeft(1).
			MarginBottom(1).
			MaxWidth(maxWidth),

		ErrorBox: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorError).
			Foreground(colorError).
			Padding(1, 2).
			MarginBottom(1).
			MaxWidth(maxWidth),

		CodeBox: lipgloss.NewStyle().
			Background(colorBgDark).
			Foreground(colorText).
			Padding(1, 2).
			MarginBottom(1).
			MaxWidth(maxWidth),

		// Text styles
		Title: lipgloss.NewStyle().
			Foreground(colorTextBright).
			Bold(true).
			MarginBottom(1),

		Subtitle: lipgloss.NewStyle().
			Foreground(colorTextDim).
			Italic(true),

		Command: lipgloss.NewStyle().
			Foreground(colorSuccess).
			Background(colorBgDark).
			Padding(0, 1),

		FilePath: lipgloss.NewStyle().
			Foreground(colorSecondary).
			Underline(true),

		Muted: lipgloss.NewStyle().
			Foreground(colorMuted),

		Emphasis: lipgloss.NewStyle().
			Foreground(colorPrimary).
			Italic(true),

		Bold: lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true),

		// Status indicators
		StatusPending: lipgloss.NewStyle().
			Foreground(colorMuted).
			SetString("○"),

		StatusInProgress: lipgloss.NewStyle().
			Foreground(colorTertiary).
			SetString("◐"),

		StatusComplete: lipgloss.NewStyle().
			Foreground(colorSuccess).
			SetString("●"),

		StatusError: lipgloss.NewStyle().
			Foreground(colorError).
			SetString("✗"),

		// Borders
		BorderActive: lipgloss.RoundedBorder(),
		BorderNormal: lipgloss.NormalBorder(),

		MaxWidth: maxWidth,
	}
}

// GlamourStyleConfig returns a custom glamour style for markdown rendering.
// Delegates to shared package for consistency across modes.
func GlamourStyleConfig() ansi.StyleConfig {
	return shared.GlamourStyleConfig()
}

// NewMarkdownRenderer creates a glamour renderer with custom styles.
func NewMarkdownRenderer() (*glamour.TermRenderer, error) {
	width := getTerminalWidth()
	if width > 120 {
		width = 120
	}

	return glamour.NewTermRenderer(
		glamour.WithStyles(GlamourStyleConfig()),
		glamour.WithWordWrap(width),
		glamour.WithEmoji(),
		glamour.WithPreservedNewLines(),
	)
}

// Icons for various UI elements - re-exported from shared package.
const (
	IconThinking   = shared.IconThinking
	IconTool       = shared.IconTool
	IconBash       = shared.IconBash
	IconSuccess    = shared.IconSuccess
	IconError      = shared.IconError
	IconWarning    = shared.IconWarning
	IconInfo       = shared.IconInfo
	IconArrowRight = shared.IconArrowRight
	IconBullet     = shared.IconBullet
	IconCheck      = shared.IconCheck
	IconCross      = shared.IconCross
	IconSpinner    = shared.IconSpinner
	IconPending    = shared.IconPending
	IconComplete   = shared.IconComplete
	IconAgent      = shared.IconAgent
	IconSubAgent   = shared.IconSubAgent
	IconEllipsis   = shared.IconEllipsis
	IconMenu       = shared.IconMenu
	IconPlan       = shared.IconPlan
)

// FormatToolLabel formats a tool label with an icon.
func FormatToolLabel(toolName string, index int) string {
	styles := DefaultStyles()
	icon := IconTool
	if toolName == "bash" {
		icon = IconBash
	}
	return styles.ToolLabel.Render(icon + " " + toolName)
}

// FormatReasoningLabel formats the reasoning section label.
func FormatReasoningLabel() string {
	styles := DefaultStyles()
	return styles.ReasoningLabel.Render(IconThinking + " Reasoning")
}

// FormatErrorLabel formats an error label.
func FormatErrorLabel(msg string) string {
	styles := DefaultStyles()
	return styles.ErrorLabel.Render(IconError + " " + msg)
}

// FormatSuccessLabel formats a success label.
func FormatSuccessLabel(msg string) string {
	styles := DefaultStyles()
	return styles.SuccessLabel.Render(IconSuccess + " " + msg)
}

// TruncateWithEllipsis truncates a string and adds ellipsis if needed.
func TruncateWithEllipsis(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// IndentText indents each line of text with the given prefix.
func IndentText(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// getTerminalWidth returns the current terminal width.
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}
