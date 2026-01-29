package shared

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Todo represents a todo item for rendering.
type Todo struct {
	Content    string `json:"content"`
	Status     string `json:"status"` // "pending", "in_progress", "completed"
	ActiveForm string `json:"active_form,omitempty"`
}

// BashParams represents bash tool parameters.
type BashParams struct {
	Command         string `json:"command"`
	Description     string `json:"description"`
	WorkingDir      string `json:"working_dir,omitempty"`
	TimeoutSeconds  int    `json:"timeout_seconds,omitempty"`
	RunInBackground bool   `json:"run_in_background,omitempty"`
}

// BashResponseMetadata represents bash tool response metadata.
type BashResponseMetadata struct {
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`
	Output      string `json:"output,omitempty"`
	ExitCode    int    `json:"exit_code,omitempty"`
	Background  bool   `json:"background,omitempty"`
	ShellID     string `json:"shell_id,omitempty"`
}

// TodosParams represents todo tool parameters.
type TodosParams struct {
	Todos []Todo `json:"todos"`
}

// TodosResponseMetadata represents todo tool response metadata.
type TodosResponseMetadata struct {
	IsNew         bool     `json:"is_new"`
	Todos         []Todo   `json:"todos"`
	JustCompleted []string `json:"just_completed,omitempty"`
	JustStarted   string   `json:"just_started,omitempty"`
	Completed     int      `json:"completed"`
	Total         int      `json:"total"`
}

// FormatTodos formats a list of todos for display.
// Uses the shared color palette for consistent styling.
func FormatTodos(todos []Todo, width int) string {
	if len(todos) == 0 {
		return ""
	}

	var lines []string

	for _, todo := range todos {
		var icon string
		var textStyle lipgloss.Style

		switch todo.Status {
		case "completed":
			icon = lipgloss.NewStyle().Foreground(ColorSuccess).Render(IconTodoCompleted)
			textStyle = lipgloss.NewStyle().Foreground(ColorTextDim).Strikethrough(true)
		case "in_progress":
			icon = lipgloss.NewStyle().Foreground(ColorPrimary).Render(IconTodoInProgress)
			textStyle = lipgloss.NewStyle().Foreground(ColorText)
		default:
			icon = lipgloss.NewStyle().Foreground(ColorMuted).Render(IconTodoPending)
			textStyle = lipgloss.NewStyle().Foreground(ColorTextDim)
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

// FormatTodoSummary returns a summary like "3/5 completed".
func FormatTodoSummary(todos []Todo) string {
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

	summaryStyle := lipgloss.NewStyle().Foreground(ColorMuted)

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

// FormatDuration formats a duration for display.
func FormatDuration(seconds float64) string {
	if seconds < 0.1 {
		return "<0.1s"
	}
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	}
	minutes := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%dm%ds", minutes, secs)
}

// TruncateText truncates text with an ellipsis if it exceeds maxLen.
func TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

// SanitizeCommand sanitizes a command string for display.
func SanitizeCommand(cmd string) string {
	cmd = strings.ReplaceAll(cmd, "\n", " ")
	cmd = strings.ReplaceAll(cmd, "\t", "    ")
	return cmd
}
