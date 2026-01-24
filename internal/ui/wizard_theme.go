package ui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// wizardTheme returns a custom theme for the agent creation wizard
// with improved spacing and styled headers.
func wizardTheme() *huh.Theme {
	t := huh.ThemeCharm()

	// Colors
	purple := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	gray := lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#94A3B8"}
	subtle := lipgloss.AdaptiveColor{Light: "#D4D4D4", Dark: "#4A4A4A"}

	// Group styles - these apply to step headers
	t.Group.Title = lipgloss.NewStyle().
		Foreground(purple).
		Bold(true).
		MarginBottom(1).
		PaddingBottom(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(subtle)

	t.Group.Description = lipgloss.NewStyle().
		Foreground(gray).
		MarginBottom(1).
		PaddingBottom(1)

	// Field separator - adds space between form fields
	t.FieldSeparator = lipgloss.NewStyle().SetString("\n\n")

	// Focused field styles
	t.Focused.Base = lipgloss.NewStyle().
		PaddingLeft(2).
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(purple)

	t.Focused.Title = lipgloss.NewStyle().
		Foreground(purple).
		Bold(true)

	t.Focused.Description = lipgloss.NewStyle().
		Foreground(gray)

	// Blurred field styles
	t.Blurred.Base = lipgloss.NewStyle().
		PaddingLeft(2).
		BorderStyle(lipgloss.HiddenBorder()).
		BorderLeft(true)

	t.Blurred.Title = lipgloss.NewStyle().
		Foreground(gray)

	t.Blurred.Description = lipgloss.NewStyle().
		Foreground(subtle)

	// Note/Card styles
	t.Focused.Card = lipgloss.NewStyle().
		PaddingLeft(2)

	t.Focused.NoteTitle = lipgloss.NewStyle().
		Foreground(purple).
		Bold(true).
		MarginBottom(1)

	// Select styles
	t.Focused.SelectSelector = lipgloss.NewStyle().
		SetString("▸ ").
		Foreground(purple)

	t.Focused.Option = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FAFAFA"})

	t.Focused.SelectedOption = lipgloss.NewStyle().
		Foreground(purple).
		Bold(true)

	// MultiSelect styles
	t.Focused.MultiSelectSelector = lipgloss.NewStyle().
		SetString("▸ ").
		Foreground(purple)

	t.Focused.SelectedPrefix = lipgloss.NewStyle().
		SetString("◉ ").
		Foreground(purple)

	t.Focused.UnselectedPrefix = lipgloss.NewStyle().
		SetString("○ ").
		Foreground(subtle)

	// Also set Blurred styles for MultiSelect
	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().
		SetString("  ")

	t.Blurred.SelectedPrefix = lipgloss.NewStyle().
		SetString("◉ ").
		Foreground(gray)

	t.Blurred.UnselectedPrefix = lipgloss.NewStyle().
		SetString("○ ").
		Foreground(subtle)

	// Button styles
	t.Focused.FocusedButton = lipgloss.NewStyle().
		Background(purple).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2).
		Bold(true)

	t.Focused.BlurredButton = lipgloss.NewStyle().
		Background(subtle).
		Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FAFAFA"}).
		Padding(0, 2)

	t.Blurred.FocusedButton = t.Focused.BlurredButton
	t.Blurred.BlurredButton = t.Focused.BlurredButton

	// Text input styles
	t.Focused.TextInput.Cursor = lipgloss.NewStyle().
		Foreground(purple)

	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().
		Foreground(subtle)

	t.Focused.TextInput.Prompt = lipgloss.NewStyle().
		Foreground(purple).
		Bold(true)

	return t
}
