package panels

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Position indicates where the sidebar is placed.
type Position int

const (
	// PositionRight places the sidebar on the right (wide terminals).
	PositionRight Position = iota
	// PositionBottom places the sidebar at the bottom (narrow terminals).
	PositionBottom
)

// MinWidthForSidebar is the minimum terminal width for right-side sidebar.
const MinWidthForSidebar = 120

// Sidebar manages the planning and memory panels.
type Sidebar struct {
	planning *PlanningPanel
	memory   *MemoryPanel

	// Layout
	position Position
	width    int
	height   int
	visible  bool

	// Active panel
	activePanel string // "planning" or "memory"

	// Keybindings
	keyMap SidebarKeyMap
}

// SidebarKeyMap defines keybindings for the sidebar.
type SidebarKeyMap struct {
	TogglePlanning key.Binding
	ToggleMemory   key.Binding
	Close          key.Binding
	ScrollUp       key.Binding
	ScrollDown     key.Binding
}

// DefaultSidebarKeyMap returns the default sidebar keybindings.
func DefaultSidebarKeyMap() SidebarKeyMap {
	return SidebarKeyMap{
		TogglePlanning: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "planning"),
		),
		ToggleMemory: key.NewBinding(
			key.WithKeys("ctrl+m"),
			key.WithHelp("ctrl+m", "memory"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close panel"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down", "scroll down"),
		),
	}
}

// NewSidebar creates a new sidebar.
func NewSidebar() *Sidebar {
	return &Sidebar{
		planning:    NewPlanningPanel(),
		memory:      NewMemoryPanel(),
		visible:     false,
		activePanel: "planning",
		keyMap:      DefaultSidebarKeyMap(),
	}
}

// SetSize updates the sidebar dimensions based on terminal size.
func (s *Sidebar) SetSize(termWidth, termHeight int) {
	// Determine position based on terminal width
	if termWidth >= MinWidthForSidebar {
		s.position = PositionRight
		s.width = termWidth / 3  // Take 1/3 of screen
		if s.width > 50 {
			s.width = 50 // Max width
		}
		if s.width < 30 {
			s.width = 30 // Min width
		}
		s.height = termHeight - 4 // Leave room for header/footer
	} else {
		s.position = PositionBottom
		s.width = termWidth - 4
		s.height = termHeight / 3 // Take 1/3 of screen
		if s.height < 8 {
			s.height = 8 // Min height
		}
	}

	// Update panel sizes
	s.planning.SetSize(s.width, s.height)
	s.memory.SetSize(s.width, s.height)
}

// Position returns the current sidebar position.
func (s *Sidebar) Position() Position {
	return s.position
}

// Width returns the sidebar width.
func (s *Sidebar) Width() int {
	if !s.visible {
		return 0
	}
	return s.width
}

// Height returns the sidebar height.
func (s *Sidebar) Height() int {
	if !s.visible {
		return 0
	}
	return s.height
}

// IsVisible returns whether the sidebar is visible.
func (s *Sidebar) IsVisible() bool {
	return s.visible
}

// ActivePanel returns the currently active panel.
func (s *Sidebar) ActivePanel() string {
	return s.activePanel
}

// TogglePlanning toggles the planning panel.
func (s *Sidebar) TogglePlanning() {
	if s.visible && s.activePanel == "planning" {
		s.visible = false
		s.planning.Hide()
	} else {
		s.visible = true
		s.activePanel = "planning"
		s.planning.Show()
		s.memory.Hide()
	}
}

// ToggleMemory toggles the memory panel.
func (s *Sidebar) ToggleMemory() {
	if s.visible && s.activePanel == "memory" {
		s.visible = false
		s.memory.Hide()
	} else {
		s.visible = true
		s.activePanel = "memory"
		s.memory.Show()
		s.planning.Hide()
	}
}

// Close hides the sidebar.
func (s *Sidebar) Close() {
	s.visible = false
	s.planning.Hide()
	s.memory.Hide()
}

// Focus gives focus to the active panel.
func (s *Sidebar) Focus() {
	if s.activePanel == "planning" {
		s.planning.Focus()
		s.memory.Blur()
	} else {
		s.memory.Focus()
		s.planning.Blur()
	}
}

// Blur removes focus from panels.
func (s *Sidebar) Blur() {
	s.planning.Blur()
	s.memory.Blur()
}

// IsFocused returns whether any panel is focused.
func (s *Sidebar) IsFocused() bool {
	return s.planning.IsFocused() || s.memory.IsFocused()
}

// SetTodos updates the planning panel todos.
func (s *Sidebar) SetTodos(todos []TodoItem) {
	s.planning.SetTodos(todos)
}

// SetMemories updates the memory panel memories.
func (s *Sidebar) SetMemories(memories []MemoryItem) {
	s.memory.SetMemories(memories)
}

// Update handles sidebar-related messages.
func (s *Sidebar) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keyMap.TogglePlanning):
			s.TogglePlanning()
			return nil
		case key.Matches(msg, s.keyMap.ToggleMemory):
			s.ToggleMemory()
			return nil
		case key.Matches(msg, s.keyMap.Close):
			if s.visible {
				s.Close()
				return nil
			}
		case key.Matches(msg, s.keyMap.ScrollUp):
			if s.IsFocused() {
				if s.activePanel == "planning" {
					s.planning.ScrollUp(1)
				} else {
					s.memory.ScrollUp(1)
				}
				return nil
			}
		case key.Matches(msg, s.keyMap.ScrollDown):
			if s.IsFocused() {
				if s.activePanel == "planning" {
					s.planning.ScrollDown(1)
				} else {
					s.memory.ScrollDown(1)
				}
				return nil
			}
		}

	case TodosUpdateMsg:
		s.SetTodos(msg.Todos)
		return nil

	case MemoriesUpdateMsg:
		s.SetMemories(msg.Memories)
		return nil
	}

	return nil
}

// View renders the sidebar.
func (s *Sidebar) View() string {
	if !s.visible {
		return ""
	}

	if s.activePanel == "planning" {
		return s.planning.View()
	}
	return s.memory.View()
}

// ContentWidth returns the available width for main content when sidebar is visible.
func (s *Sidebar) ContentWidth(termWidth int) int {
	if !s.visible || s.position == PositionBottom {
		return termWidth
	}
	return termWidth - s.width - 1 // -1 for separator
}

// ContentHeight returns the available height for main content when sidebar is visible.
func (s *Sidebar) ContentHeight(termHeight int) int {
	if !s.visible || s.position == PositionRight {
		return termHeight
	}
	return termHeight - s.height - 1 // -1 for separator
}

// RenderWithContent renders the sidebar alongside the main content.
func (s *Sidebar) RenderWithContent(content string, termWidth, termHeight int) string {
	if !s.visible {
		return content
	}

	sidebarView := s.View()

	if s.position == PositionRight {
		// Side by side
		separator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151")).
			Render("│")

		// Ensure content fills available width
		contentWidth := termWidth - s.width - 1
		contentStyle := lipgloss.NewStyle().Width(contentWidth)

		return lipgloss.JoinHorizontal(lipgloss.Top,
			contentStyle.Render(content),
			separator,
			sidebarView,
		)
	}

	// Stacked vertically
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#374151")).
		Width(termWidth).
		Render("─")

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		separator,
		sidebarView,
	)
}
