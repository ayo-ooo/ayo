package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/alexcabrera/ayo/internal/session"
)

// Plan item status icons.
const (
	IconPlanCompleted  = "✓"
	IconPlanInProgress = "▸"
	IconPlanPending    = "○"
	IconPlanPhase      = "■"
)

// FormatPlan formats a hierarchical plan for CLI display.
func FormatPlan(plan session.Plan, width int) string {
	if plan.IsEmpty() {
		return ""
	}

	var lines []string

	if plan.IsFlat() {
		// Just tasks (no phases)
		for _, task := range plan.Tasks {
			lines = append(lines, formatTask(task, width, 0)...)
		}
	} else {
		// Phases with tasks
		for _, phase := range plan.Phases {
			lines = append(lines, formatPhase(phase, width)...)
		}
	}

	return strings.Join(lines, "\n")
}

func formatPhase(phase session.Phase, width int) []string {
	var lines []string

	// Phase header
	phaseStyle := lipgloss.NewStyle().Bold(true)
	var icon string
	var nameStyle lipgloss.Style

	switch phase.Status {
	case session.PlanStatusCompleted:
		icon = lipgloss.NewStyle().Foreground(colorSuccess).Render(IconPlanCompleted)
		nameStyle = phaseStyle.Foreground(colorTextDim)
	case session.PlanStatusInProgress:
		icon = lipgloss.NewStyle().Foreground(colorPrimary).Render(IconPlanPhase)
		nameStyle = phaseStyle.Foreground(colorText)
	default:
		icon = lipgloss.NewStyle().Foreground(colorMuted).Render(IconPlanPhase)
		nameStyle = phaseStyle.Foreground(colorTextDim)
	}

	lines = append(lines, fmt.Sprintf("%s %s", icon, nameStyle.Render(phase.Name)))

	// Tasks within phase (indented)
	for _, task := range phase.Tasks {
		lines = append(lines, formatTask(task, width, 1)...)
	}

	return lines
}

func formatTask(task session.Task, width int, indent int) []string {
	var lines []string
	indentStr := strings.Repeat("  ", indent)

	var icon string
	var textStyle lipgloss.Style

	switch task.Status {
	case session.PlanStatusCompleted:
		icon = lipgloss.NewStyle().Foreground(colorSuccess).Render(IconPlanCompleted)
		textStyle = lipgloss.NewStyle().Foreground(colorTextDim).Strikethrough(true)
	case session.PlanStatusInProgress:
		icon = lipgloss.NewStyle().Foreground(colorPrimary).Render(IconPlanInProgress)
		textStyle = lipgloss.NewStyle().Foreground(colorText)
	default:
		icon = lipgloss.NewStyle().Foreground(colorMuted).Render(IconPlanPending)
		textStyle = lipgloss.NewStyle().Foreground(colorTextDim)
	}

	// Use active_form for in-progress, content otherwise
	text := task.Content
	if task.Status == session.PlanStatusInProgress && task.ActiveForm != "" {
		text = task.ActiveForm
	}

	// Truncate if needed
	maxTextWidth := width - 4 - (indent * 2)
	if maxTextWidth > 0 && len(text) > maxTextWidth {
		text = text[:maxTextWidth-1] + "…"
	}

	lines = append(lines, fmt.Sprintf("%s%s %s", indentStr, icon, textStyle.Render(text)))

	// Todos within task (further indented)
	for _, todo := range task.Todos {
		lines = append(lines, formatTodo(todo, width, indent+1))
	}

	return lines
}

func formatTodo(todo session.Todo, width int, indent int) string {
	indentStr := strings.Repeat("  ", indent)

	var icon string
	var textStyle lipgloss.Style

	switch todo.Status {
	case session.PlanStatusCompleted:
		icon = lipgloss.NewStyle().Foreground(colorSuccess).Render(IconPlanCompleted)
		textStyle = lipgloss.NewStyle().Foreground(colorTextDim).Strikethrough(true)
	case session.PlanStatusInProgress:
		icon = lipgloss.NewStyle().Foreground(colorPrimary).Render(IconPlanInProgress)
		textStyle = lipgloss.NewStyle().Foreground(colorText)
	default:
		icon = lipgloss.NewStyle().Foreground(colorMuted).Render(IconPlanPending)
		textStyle = lipgloss.NewStyle().Foreground(colorTextDim)
	}

	text := todo.Content
	if todo.Status == session.PlanStatusInProgress && todo.ActiveForm != "" {
		text = todo.ActiveForm
	}

	maxTextWidth := width - 4 - (indent * 2)
	if maxTextWidth > 0 && len(text) > maxTextWidth {
		text = text[:maxTextWidth-1] + "…"
	}

	return fmt.Sprintf("%s%s %s", indentStr, icon, textStyle.Render(text))
}

// FormatPlanSummary returns a summary like "3/5 completed".
func FormatPlanSummary(plan session.Plan) string {
	pending, inProgress, completed := plan.Stats()
	total := pending + inProgress + completed
	return fmt.Sprintf("%d/%d completed", completed, total)
}

// FormatPlanChange returns a description of what changed.
func FormatPlanChange(isNew bool, justCompleted []string, justStarted string, completed, total int) string {
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
		return summaryStyle.Render("starting task")
	}

	return fmt.Sprintf("%d/%d", completed, total)
}

// Legacy function for backward compatibility
// FormatPlanList formats a flat task list (deprecated, use FormatPlan).
func FormatPlanList(tasks []session.Task, width int) string {
	if len(tasks) == 0 {
		return ""
	}
	plan := session.Plan{Tasks: tasks}
	return FormatPlan(plan, width)
}
