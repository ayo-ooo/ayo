package shared

import (
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

// rendererCache stores glamour renderers by width to avoid recreating them.
var rendererCache sync.Map // map[int]*glamour.TermRenderer

// GetMarkdownRenderer returns a cached glamour renderer for the given width.
// Width is clamped to 40-200 to prevent cache explosion.
// Uses auto-style which adapts to terminal background.
func GetMarkdownRenderer(width int) *glamour.TermRenderer {
	width = Clamp(width, 40, 200)

	if r, ok := rendererCache.Load(width); ok {
		return r.(*glamour.TermRenderer)
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}

	rendererCache.Store(width, r)
	return r
}

// styledRendererCache stores glamour renderers with custom styling by width.
var styledRendererCache sync.Map // map[int]*glamour.TermRenderer

// GetStyledMarkdownRenderer returns a cached glamour renderer with custom styling.
// Uses the ayo color palette for consistent theming.
func GetStyledMarkdownRenderer(width int) *glamour.TermRenderer {
	width = Clamp(width, 40, 200)

	if r, ok := styledRendererCache.Load(width); ok {
		return r.(*glamour.TermRenderer)
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(GlamourStyleConfig()),
		glamour.WithWordWrap(width),
		glamour.WithEmoji(),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		return nil
	}

	styledRendererCache.Store(width, r)
	return r
}

// ClearRendererCache clears all renderer caches (useful for testing).
func ClearRendererCache() {
	rendererCache.Range(func(key, value any) bool {
		rendererCache.Delete(key)
		return true
	})
	styledRendererCache.Range(func(key, value any) bool {
		styledRendererCache.Delete(key)
		return true
	})
}

// Clamp constrains a value to a range.
func Clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// GlamourStyleConfig returns the custom glamour style configuration.
// Uses the ayo color palette for consistent theming across the application.
func GlamourStyleConfig() ansi.StyleConfig {
	margin := uint(0)
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
					Color: strPtr("#67e8f9"),
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
			BlockPrefix: "\n  ",
		},
	}
}

// Helper functions for pointer creation
func strPtr(s string) *string   { return &s }
func boolPtr(b bool) *bool      { return &b }
func uintPtr(u uint) *uint      { return &u }
