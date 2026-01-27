// Package panels provides expandable side panels for the chat TUI.
package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// TodoItem represents a single task in the planning panel.
type TodoItem struct {
	Content    string
	ActiveForm string
	Status     string // "pending", "in_progress", "completed"
}

// PlanningPanel displays the current todo list.
type PlanningPanel struct {
	viewport viewport.Model
	todos    []TodoItem
	width    int
	height   int
	visible  bool
	focused  bool
}

// NewPlanningPanel creates a new planning panel.
func NewPlanningPanel() *PlanningPanel {
	return &PlanningPanel{
		todos:   make([]TodoItem, 0),
		visible: false,
	}
}

// SetSize updates the panel dimensions.
func (p *PlanningPanel) SetSize(width, height int) {
	p.width = width
	p.height = height

	if p.viewport.Width == 0 {
		p.viewport = viewport.New(width-4, height-4) // Account for borders
		p.viewport.MouseWheelEnabled = true
	} else {
		p.viewport.Width = width - 4
		p.viewport.Height = height - 4
	}

	p.updateContent()
}

// SetTodos updates the todo list.
func (p *PlanningPanel) SetTodos(todos []TodoItem) {
	p.todos = todos
	p.updateContent()
}

// Toggle shows or hides the panel.
func (p *PlanningPanel) Toggle() {
	p.visible = !p.visible
}

// Show makes the panel visible.
func (p *PlanningPanel) Show() {
	p.visible = true
}

// Hide hides the panel.
func (p *PlanningPanel) Hide() {
	p.visible = false
}

// IsVisible returns whether the panel is visible.
func (p *PlanningPanel) IsVisible() bool {
	return p.visible
}

// Focus gives focus to the panel.
func (p *PlanningPanel) Focus() {
	p.focused = true
}

// Blur removes focus from the panel.
func (p *PlanningPanel) Blur() {
	p.focused = false
}

// IsFocused returns whether the panel is focused.
func (p *PlanningPanel) IsFocused() bool {
	return p.focused
}

// View renders the panel.
func (p *PlanningPanel) View() string {
	if !p.visible || p.width <= 0 || p.height <= 0 {
		return ""
	}

	// Styles
	borderColor := lipgloss.Color("#6b7280")
	if p.focused {
		borderColor = lipgloss.Color("#a78bfa")
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#a78bfa"))

	containerStyle := lipgloss.NewStyle().
		Width(p.width).
		Height(p.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	// Header
	title := titleStyle.Render("Planning")
	stats := p.renderStats()
	header := lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", stats)

	// Content
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		p.viewport.View(),
	)

	return containerStyle.Render(content)
}

// updateContent rebuilds the viewport content from todos.
func (p *PlanningPanel) updateContent() {
	if len(p.todos) == 0 {
		p.viewport.SetContent(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")).
			Render("No tasks"))
		return
	}

	var lines []string
	for _, todo := range p.todos {
		line := p.renderTodoItem(todo)
		lines = append(lines, line)
	}

	p.viewport.SetContent(strings.Join(lines, "\n"))
}

// renderTodoItem renders a single todo item.
func (p *PlanningPanel) renderTodoItem(todo TodoItem) string {
	// Status icons
	var icon string
	var textStyle lipgloss.Style

	switch todo.Status {
	case "completed":
		icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Render("✓")
		textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Strikethrough(true)
	case "in_progress":
		icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#3b82f6")).Render("▸")
		textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e5e7eb"))
	default: // pending
		icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("○")
		textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af"))
	}

	// Use active form if in progress, otherwise use content
	text := todo.Content
	if todo.Status == "in_progress" && todo.ActiveForm != "" {
		text = todo.ActiveForm
	}

	// Truncate if needed
	maxWidth := p.width - 8 // Account for icon and padding
	if len(text) > maxWidth && maxWidth > 3 {
		text = text[:maxWidth-3] + "..."
	}

	return fmt.Sprintf("%s %s", icon, textStyle.Render(text))
}

// renderStats returns the progress statistics.
func (p *PlanningPanel) renderStats() string {
	if len(p.todos) == 0 {
		return ""
	}

	var completed, inProgress int
	for _, todo := range p.todos {
		switch todo.Status {
		case "completed":
			completed++
		case "in_progress":
			inProgress++
		}
	}

	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	return statsStyle.Render(fmt.Sprintf("%d/%d", completed, len(p.todos)))
}

// ScrollUp scrolls the viewport up.
func (p *PlanningPanel) ScrollUp(n int) {
	p.viewport.LineUp(n)
}

// ScrollDown scrolls the viewport down.
func (p *PlanningPanel) ScrollDown(n int) {
	p.viewport.LineDown(n)
}

// TodosUpdateMsg is sent when the todo list changes.
type TodosUpdateMsg struct {
	Todos []TodoItem
}
