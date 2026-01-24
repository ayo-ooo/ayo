package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Color palette - a cohesive dark theme inspired by popular terminal themes.
var (
	// Primary colors
	colorPrimary     = lipgloss.Color("#a78bfa") // Purple - main accent
	colorSecondary   = lipgloss.Color("#67e8f9") // Cyan - secondary accent
	colorTertiary    = lipgloss.Color("#fbbf24") // Amber - warnings/tool labels
	colorSuccess     = lipgloss.Color("#4ade80") // Green - success states
	colorError       = lipgloss.Color("#f87171") // Red - errors
	colorMuted       = lipgloss.Color("#6b7280") // Gray - muted text
	colorSubtle      = lipgloss.Color("#374151") // Dark gray - borders/backgrounds

	// Text colors
	colorText       = lipgloss.Color("#e5e7eb") // Light gray - main text
	colorTextDim    = lipgloss.Color("#9ca3af") // Medium gray - dim text
	colorTextBright = lipgloss.Color("#f9fafb") // White - bright text

	// Background colors
	colorBgDark     = lipgloss.Color("#1f2937") // Dark background
	colorBgSubtle   = lipgloss.Color("#111827") // Darker background
	colorBgAccent   = lipgloss.Color("#312e81") // Purple tinted background
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
func GlamourStyleConfig() ansi.StyleConfig {
	margin := uint(2)

	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockPrefix: "",
				BlockSuffix: "",
				Color:       strPtr("#e5e7eb"),
			},
			Margin: &margin,
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr("#67e8f9"),
				Italic: boolPtr(true),
			},
			Indent:      uintPtr(1),
			IndentToken: strPtr("│ "),
		},
		List: ansi.StyleList{
			LevelIndent: 2,
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: strPtr("#e5e7eb"),
				},
			},
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       strPtr("#a78bfa"),
				Bold:        boolPtr(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "# ",
				Color:  strPtr("#a78bfa"),
				Bold:   boolPtr(true),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "## ",
				Color:  strPtr("#a78bfa"),
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "### ",
				Color:  strPtr("#c4b5fd"),
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "#### ",
				Color:  strPtr("#c4b5fd"),
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "##### ",
				Color:  strPtr("#ddd6fe"),
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "###### ",
				Color:  strPtr("#ddd6fe"),
			},
		},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: ansi.StylePrimitive{
			Color:  strPtr("#fbbf24"),
			Italic: boolPtr(true),
		},
		Strong: ansi.StylePrimitive{
			Bold:  boolPtr(true),
			Color: strPtr("#f9fafb"),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  strPtr("#374151"),
			Format: "\n────────────────────────────────\n",
		},
		Item: ansi.StylePrimitive{
			BlockPrefix: "- ",
		},
		Enumeration: ansi.StylePrimitive{
			BlockPrefix: ". ",
			Color:       strPtr("#67e8f9"),
		},
		Task: ansi.StyleTask{
			StylePrimitive: ansi.StylePrimitive{},
			Ticked:         "[x] ",
			Unticked:       "[ ] ",
		},
		Link: ansi.StylePrimitive{
			Color:     strPtr("#67e8f9"),
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: strPtr("#a78bfa"),
		},
		Image: ansi.StylePrimitive{
			Color:     strPtr("#67e8f9"),
			Underline: boolPtr(true),
		},
		ImageText: ansi.StylePrimitive{
			Color:  strPtr("#a78bfa"),
			Format: "Image: {{.text}}",
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:           strPtr("#4ade80"),
				BackgroundColor: strPtr("#1f2937"),
				Prefix:          " ",
				Suffix:          " ",
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: strPtr("#e5e7eb"),
				},
				Margin: uintPtr(2),
			},
			Chroma: &ansi.Chroma{
				Text: ansi.StylePrimitive{
					Color: strPtr("#e5e7eb"),
				},
				Error: ansi.StylePrimitive{
					Color:           strPtr("#f9fafb"),
					BackgroundColor: strPtr("#f87171"),
				},
				Comment: ansi.StylePrimitive{
					Color: strPtr("#6b7280"),
				},
				CommentPreproc: ansi.StylePrimitive{
					Color: strPtr("#fbbf24"),
				},
				Keyword: ansi.StylePrimitive{
					Color: strPtr("#a78bfa"),
				},
				KeywordReserved: ansi.StylePrimitive{
					Color: strPtr("#c084fc"),
				},
				KeywordNamespace: ansi.StylePrimitive{
					Color: strPtr("#f472b6"),
				},
				KeywordType: ansi.StylePrimitive{
					Color: strPtr("#67e8f9"),
				},
				Operator: ansi.StylePrimitive{
					Color: strPtr("#f87171"),
				},
				Punctuation: ansi.StylePrimitive{
					Color: strPtr("#9ca3af"),
				},
				Name: ansi.StylePrimitive{
					Color: strPtr("#e5e7eb"),
				},
				NameBuiltin: ansi.StylePrimitive{
					Color: strPtr("#67e8f9"),
				},
				NameTag: ansi.StylePrimitive{
					Color: strPtr("#a78bfa"),
				},
				NameAttribute: ansi.StylePrimitive{
					Color: strPtr("#4ade80"),
				},
				NameClass: ansi.StylePrimitive{
					Color:     strPtr("#f9fafb"),
					Underline: boolPtr(true),
					Bold:      boolPtr(true),
				},
				NameConstant: ansi.StylePrimitive{
					Color: strPtr("#c084fc"),
				},
				NameDecorator: ansi.StylePrimitive{
					Color: strPtr("#fbbf24"),
				},
				NameFunction: ansi.StylePrimitive{
					Color: strPtr("#4ade80"),
				},
				LiteralNumber: ansi.StylePrimitive{
					Color: strPtr("#67e8f9"),
				},
				LiteralString: ansi.StylePrimitive{
					Color: strPtr("#fbbf24"),
				},
				LiteralStringEscape: ansi.StylePrimitive{
					Color: strPtr("#f472b6"),
				},
				GenericDeleted: ansi.StylePrimitive{
					Color: strPtr("#f87171"),
				},
				GenericEmph: ansi.StylePrimitive{
					Italic: boolPtr(true),
				},
				GenericInserted: ansi.StylePrimitive{
					Color: strPtr("#4ade80"),
				},
				GenericStrong: ansi.StylePrimitive{
					Bold: boolPtr(true),
				},
				GenericSubheading: ansi.StylePrimitive{
					Color: strPtr("#a78bfa"),
				},
				Background: ansi.StylePrimitive{
					BackgroundColor: strPtr("#111827"),
				},
			},
		},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{},
			},
			CenterSeparator: strPtr("┼"),
			ColumnSeparator: strPtr("│"),
			RowSeparator:    strPtr("─"),
		},
		DefinitionDescription: ansi.StylePrimitive{
			BlockPrefix: "\n> ",
		},
	}
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

// Icons for various UI elements - colorizable Unicode glyphs (no emojis).
// These are carefully selected from box drawing, geometric shapes, and
// miscellaneous symbols to be visually distinct and terminal-compatible.
const (
	IconThinking   = "◇"  // White diamond - thinking/reasoning
	IconTool       = "▶"  // Black right-pointing triangle - tool execution
	IconBash       = "❯"  // Heavy right angle bracket - bash/shell prompt
	IconSuccess    = "✓"  // Check mark - success
	IconError      = "✗"  // Ballot X - error
	IconWarning    = "△"  // White up-pointing triangle - warning
	IconInfo       = "●"  // Black circle - info
	IconArrowRight = "→"  // Rightwards arrow - navigation
	IconBullet     = "•"  // Bullet - list items
	IconCheck      = "✓"  // Check mark - completed
	IconCross      = "✗"  // Ballot X - failed
	IconSpinner    = "◐"  // Circle with left half black - in progress
	IconPending    = "○"  // White circle - pending
	IconComplete   = "●"  // Black circle - complete
	IconAgent      = "◆"  // Black diamond - agent
	IconSubAgent   = "▹"  // White right-pointing small triangle - sub-agent
	IconEllipsis   = "⋯"  // Midline horizontal ellipsis - loading/truncated
	IconMenu       = "≡"  // Identical to - menu (hamburger alternative)
	IconPlan       = "□"  // White square - plan/todo
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

// Helper functions for pointer types.
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func uintPtr(u uint) *uint    { return &u }
