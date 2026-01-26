package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Todo item status icons.
const (
	IconTodoCompleted  = "✓"
	IconTodoInProgress = "▸"
	IconTodoPending    = "○"
)

// UITodo represents a todo item for UI rendering.
// This mirrors run.Todo to avoid import cycles.
type UITodo struct {
	Content    string `json:"content"`
	Status     string `json:"status"` // "pending", "in_progress", "completed"
	ActiveForm string `json:"active_form"`
}

// FormatUITodos formats a flat todo list for CLI display.
func FormatUITodos(todos []UITodo, width int) string {
	if len(todos) == 0 {
		return ""
	}

	var lines []string

	for _, todo := range todos {
		var icon string
		var textStyle lipgloss.Style

		switch todo.Status {
		case "completed":
			icon = lipgloss.NewStyle().Foreground(colorSuccess).Render(IconTodoCompleted)
			textStyle = lipgloss.NewStyle().Foreground(colorTextDim).Strikethrough(true)
		case "in_progress":
			icon = lipgloss.NewStyle().Foreground(colorPrimary).Render(IconTodoInProgress)
			textStyle = lipgloss.NewStyle().Foreground(colorText)
		default:
			icon = lipgloss.NewStyle().Foreground(colorMuted).Render(IconTodoPending)
			textStyle = lipgloss.NewStyle().Foreground(colorTextDim)
		}

		// Use active_form for in-progress, content otherwise
		text := todo.Content
		if todo.Status == "in_progress" && todo.ActiveForm != "" {
			text = todo.ActiveForm
		}

		// Truncate if needed
		maxTextWidth := width - 4
		if maxTextWidth > 0 && len(text) > maxTextWidth {
			text = text[:maxTextWidth-1] + "..."
		}

		lines = append(lines, fmt.Sprintf("%s %s", icon, textStyle.Render(text)))
	}

	return strings.Join(lines, "\n")
}

// FormatUITodoSummary returns a summary like "3/5 completed".
func FormatUITodoSummary(todos []UITodo) string {
	pending, inProgress, completed := 0, 0, 0
	for _, todo := range todos {
		switch todo.Status {
		case "pending":
			pending++
		case "in_progress":
			inProgress++
		case "completed":
			completed++
		}
	}
	total := pending + inProgress + completed
	return fmt.Sprintf("%d/%d completed", completed, total)
}

// FormatTodoChange returns a description of what changed.
func FormatTodoChange(isNew bool, justCompleted []string, justStarted string, completed, total int) string {
	if isNew {
		if justStarted != "" {
			return fmt.Sprintf("created %d items, starting first", total)
		}
		return fmt.Sprintf("created %d items", total)
	}

	hasCompleted := len(justCompleted) > 0
	hasStarted := justStarted != ""
	allCompleted := completed == total

	summaryStyle := lipgloss.NewStyle().Foreground(colorMuted)

	if allCompleted {
		return summaryStyle.Render("completed all items")
	}

	if hasCompleted && hasStarted {
		return summaryStyle.Render(fmt.Sprintf("completed %d, starting next", len(justCompleted)))
	}
	if hasCompleted {
		return summaryStyle.Render(fmt.Sprintf("completed %d", len(justCompleted)))
	}
	if hasStarted {
		return summaryStyle.Render("starting todo")
	}

	return fmt.Sprintf("%d/%d", completed, total)
}
