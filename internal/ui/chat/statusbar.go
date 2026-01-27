package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar displays memory count, task progress, and keyboard hints.
type StatusBar struct {
	width int

	// Memory state
	memoryCount int

	// Task state
	currentTask   string
	completedTasks int
	totalTasks    int

	// Keyboard hints based on current focus
	hints string
}

// NewStatusBar creates a new status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{
		hints: "enter send · shift+enter newline · ctrl+c quit",
	}
}

// SetWidth sets the status bar width.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetMemoryCount updates the memory count display.
func (s *StatusBar) SetMemoryCount(count int) {
	s.memoryCount = count
}

// SetTaskProgress updates the task progress display.
func (s *StatusBar) SetTaskProgress(current string, completed, total int) {
	s.currentTask = current
	s.completedTasks = completed
	s.totalTasks = total
}

// SetHints updates the keyboard hints.
func (s *StatusBar) SetHints(hints string) {
	s.hints = hints
}

// Render returns the status bar string.
func (s *StatusBar) Render() string {
	if s.width <= 0 {
		s.width = 80
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#a78bfa"))

	var parts []string

	// Memory indicator
	if s.memoryCount > 0 {
		memIcon := highlightStyle.Render("◆")
		memText := style.Render(fmt.Sprintf(" %d memories", s.memoryCount))
		parts = append(parts, memIcon+memText)
	}

	// Task progress
	if s.totalTasks > 0 {
		progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3b82f6"))
		progress := progressStyle.Render(fmt.Sprintf("%d/%d", s.completedTasks, s.totalTasks))

		if s.currentTask != "" {
			taskStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e"))
			arrow := taskStyle.Render("▸")
			task := style.Render(" " + s.truncateTask(s.currentTask, 30))
			parts = append(parts, progress+" "+arrow+task)
		} else {
			parts = append(parts, progress)
		}
	}

	left := strings.Join(parts, " · ")

	// Right side: hints
	right := style.Render(s.hints)

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := s.width - leftWidth - rightWidth - 4 // -4 for padding

	if spacing < 1 {
		// Not enough space, just show hints
		return style.Render("  " + s.hints)
	}

	spacer := strings.Repeat(" ", spacing)
	return "  " + left + spacer + right
}

// truncateTask truncates the task text to maxLen.
func (s *StatusBar) truncateTask(task string, maxLen int) string {
	if len(task) <= maxLen {
		return task
	}
	return task[:maxLen-3] + "..."
}

// Update processes status bar messages.
func (s *StatusBar) Update(msg interface{}) {
	switch m := msg.(type) {
	case MemoryCountMsg:
		s.SetMemoryCount(m.Count)
	case TaskProgressMsg:
		s.SetTaskProgress(m.Current, m.Completed, m.Total)
	case HintsMsg:
		s.SetHints(m.Hints)
	}
}

// MemoryCountMsg updates the memory count.
type MemoryCountMsg struct {
	Count int
}

// TaskProgressMsg updates task progress.
type TaskProgressMsg struct {
	Current   string
	Completed int
	Total     int
}

// HintsMsg updates keyboard hints.
type HintsMsg struct {
	Hints string
}
