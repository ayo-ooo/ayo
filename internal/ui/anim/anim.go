// Package anim provides loading animation components for the TUI.
package anim

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Settings configures the animation appearance.
type Settings struct {
	// Size is the width of the animation in characters.
	Size int
	// Label is optional text displayed alongside the animation.
	Label string
	// GradColorA is the primary gradient color.
	GradColorA lipgloss.Color
	// GradColorB is the secondary gradient color.
	GradColorB lipgloss.Color
	// LabelColor is the color for the label text.
	LabelColor lipgloss.Color
	// CycleColors enables color cycling.
	CycleColors bool
	// Interval is the animation tick interval.
	Interval time.Duration
}

// DefaultSettings returns sensible default animation settings.
func DefaultSettings() Settings {
	return Settings{
		Size:        15,
		Label:       "Working",
		GradColorA:  lipgloss.Color("#a78bfa"),
		GradColorB:  lipgloss.Color("#67e8f9"),
		LabelColor:  lipgloss.Color("#9ca3af"),
		CycleColors: true,
		Interval:    80 * time.Millisecond,
	}
}

// TickMsg triggers animation frame advancement.
type TickMsg struct {
	ID string
}

// Model is the animation component.
type Model struct {
	id       string
	settings Settings
	frame    int
	active   bool
}

// Spinner frames (braille dots).
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// New creates a new animation model.
func New(settings Settings) Model {
	if settings.Interval == 0 {
		settings.Interval = 80 * time.Millisecond
	}
	if settings.Size == 0 {
		settings.Size = 15
	}
	return Model{
		id:       generateID(),
		settings: settings,
		active:   true,
	}
}

// ID returns the animation's unique identifier.
func (m Model) ID() string {
	return m.id
}

// generateID creates a simple unique ID.
func generateID() string {
	return time.Now().Format("20060102150405.000000")
}

// Init starts the animation ticker.
func (m Model) Init() tea.Cmd {
	return m.tick()
}

// tick returns a command that sends a tick message after the interval.
func (m Model) tick() tea.Cmd {
	return tea.Tick(m.settings.Interval, func(t time.Time) tea.Msg {
		return TickMsg{ID: m.id}
	})
}

// Update handles tick messages to advance the animation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if msg.ID != m.id {
			return m, nil
		}
		if !m.active {
			return m, nil
		}
		m.frame = (m.frame + 1) % len(spinnerFrames)
		return m, m.tick()
	}
	return m, nil
}

// View renders the animation.
func (m Model) View() string {
	if !m.active {
		return ""
	}

	spinnerStyle := lipgloss.NewStyle().Foreground(m.settings.GradColorA)
	spinner := spinnerStyle.Render(spinnerFrames[m.frame])

	if m.settings.Label == "" {
		return spinner
	}

	labelStyle := lipgloss.NewStyle().Foreground(m.settings.LabelColor)
	return spinner + " " + labelStyle.Render(m.settings.Label)
}

// Start activates the animation.
func (m *Model) Start() tea.Cmd {
	m.active = true
	return m.tick()
}

// Stop deactivates the animation.
func (m *Model) Stop() {
	m.active = false
}

// IsActive returns whether the animation is running.
func (m Model) IsActive() bool {
	return m.active
}

// SetLabel updates the animation label.
func (m *Model) SetLabel(label string) {
	m.settings.Label = label
}
