package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	// "github.com/alexcabrera/ayo/internal/session" - Removed as part of framework cleanup
)

// Message represents a simple message for history rendering
// Replaces session.Message for build system compatibility
type Message struct {
	Role     string    `json:"role"`
	Content  string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Role constants for message types
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// RenderHistory converts messages into a styled string for display in the viewport.
func RenderHistory(messages []Message, agentHandle string) string {
	if len(messages) == 0 {
		return renderEmptyHistory()
	}

	var parts []string

	for _, msg := range messages {
		// Skip system messages
		if msg.Role == RoleSystem {
			continue
		}

		rendered := renderMessage(msg, agentHandle)
		if rendered != "" {
			parts = append(parts, rendered)
		}
	}

	if len(parts) == 0 {
		return renderEmptyHistory()
	}

	return strings.Join(parts, "\n\n")
}

// RenderHistoryPreview renders a compact preview of the last N messages.
// Used for showing context when continuing a session.
func RenderHistoryPreview(messages []Message, agentHandle string, count int) string {
	if len(messages) == 0 {
		return ""
	}

	// Filter out system messages first
	var filtered []Message
	for _, msg := range messages {
		if msg.Role != RoleSystem {
			filtered = append(filtered, msg)
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	// Take last 'count' messages
	start := max(0, len(filtered)-count)
	previewMessages := filtered[start:]

	var parts []string
	for _, msg := range previewMessages {
		rendered := renderMessage(msg, agentHandle)
		if rendered != "" {
			parts = append(parts, rendered)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n")
}

func renderEmptyHistory() string {
	return lipgloss.NewStyle().Faint(true).Render("(no history)")
}

func renderMessage(msg Message, agentHandle string) string {
	switch msg.Role {
	case RoleUser:
		return renderUserMessage(msg.Content)
	case RoleAssistant:
		return renderAssistantMessage(msg.Content, agentHandle)
	case RoleTool:
		return renderToolMessage(msg.Content)
	default:
		return ""
	}
}

func renderUserMessage(content string) string {
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5865F2")).Bold(true)
	contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5865F2"))
	return fmt.Sprintf("%s\n%s", userStyle.Render("User:"), contentStyle.Render(content))
}

func renderAssistantMessage(content string, agentHandle string) string {
	agentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8ED1FC")).Bold(true)
	contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8ED1FC"))
	return fmt.Sprintf("%s\n%s", agentStyle.Render(agentHandle+":"), contentStyle.Render(content))
}

func renderToolMessage(content string) string {
	toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FEE75C")).Bold(true)
	contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FEE75C"))
	return fmt.Sprintf("%s\n%s", toolStyle.Render("Tool:"), contentStyle.Render(content))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
