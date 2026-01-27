// Package layout provides core interfaces for composable TUI components.
// These interfaces follow patterns from Crush and enable consistent
// focus management, sizing, and positioning across all chat components.
package layout

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Focusable defines components that can receive and lose focus.
// Components should visually indicate focus state and may change
// behavior (e.g., accepting keyboard input) based on focus.
type Focusable interface {
	// Focus gives the component focus. Returns a command that may
	// be used to start cursor blink or other focus-related effects.
	Focus() tea.Cmd

	// Blur removes focus from the component. Returns a command that
	// may be used to stop animations or clean up focus state.
	Blur() tea.Cmd

	// IsFocused returns whether the component currently has focus.
	IsFocused() bool
}

// Sizeable defines components that have configurable dimensions.
// Components should adapt their rendering to fit the given size.
type Sizeable interface {
	// SetSize updates the component's dimensions. Returns a command
	// that may be used to trigger re-rendering or layout updates.
	SetSize(width, height int) tea.Cmd

	// GetSize returns the component's current dimensions.
	GetSize() (width, height int)
}

// Positional defines components that can be positioned within a parent.
// This is useful for absolute positioning or overlay components.
type Positional interface {
	// SetPosition updates the component's position. Returns a command
	// that may be used to trigger re-rendering.
	SetPosition(x, y int) tea.Cmd
}

// Help defines components that provide keybinding documentation.
// This enables consistent help display across the application.
type Help interface {
	// Bindings returns the keybindings this component responds to.
	// Used to build help text and keyboard shortcut documentation.
	Bindings() []key.Binding
}

// Model extends tea.Model with a typed Update method that returns
// the concrete Model type instead of tea.Model. This enables
// better type safety when composing components.
type Model interface {
	// Init initializes the component and returns an initial command.
	Init() tea.Cmd

	// Update handles a message and returns the updated model and command.
	Update(msg tea.Msg) (Model, tea.Cmd)

	// View renders the component to a string.
	View() string
}
