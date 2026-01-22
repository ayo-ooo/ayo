package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/alexcabrera/ayo/internal/session"
)

// RenderHistory converts session messages into a styled string for display in the viewport.
func RenderHistory(messages []session.Message, agentHandle string) string {
	if len(messages) == 0 {
		return renderEmptyHistory()
	}

	var parts []string

	for _, msg := range messages {
		// Skip system messages
		if msg.Role == session.RoleSystem {
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
func RenderHistoryPreview(messages []session.Message, agentHandle string, count int) string {
	if len(messages) == 0 {
		return ""
	}

	// Filter out system messages first
	var filtered []session.Message
	for _, msg := range messages {
		if msg.Role != session.RoleSystem {
			filtered = append(filtered, msg)
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	// Take last N messages
	start := 0
	if len(filtered) > count {
		start = len(filtered) - count
	}
	recent := filtered[start:]

	var parts []string
	for _, msg := range recent {
		rendered := renderMessageCompact(msg, agentHandle)
		if rendered != "" {
			parts = append(parts, rendered)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	// Add header
	headerStyle := lipgloss.NewStyle().Foreground(colorMuted).Italic(true)
	header := headerStyle.Render(fmt.Sprintf("─── Last %d messages (^H for full history) ───", len(recent)))

	return header + "\n\n" + strings.Join(parts, "\n\n")
}

func renderEmptyHistory() string {
	style := lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true)
	return style.Render("No conversation history to display.")
}

func renderMessage(msg session.Message, agentHandle string) string {
	switch msg.Role {
	case session.RoleUser:
		return renderUserMessage(msg)
	case session.RoleAssistant:
		return renderAssistantMessage(msg, agentHandle)
	case session.RoleTool:
		return renderToolMessage(msg)
	default:
		return ""
	}
}

// renderMessageCompact renders a more compact version for preview.
func renderMessageCompact(msg session.Message, agentHandle string) string {
	switch msg.Role {
	case session.RoleUser:
		return renderUserMessageCompact(msg)
	case session.RoleAssistant:
		return renderAssistantMessageCompact(msg, agentHandle)
	case session.RoleTool:
		// Skip tool messages in compact view
		return ""
	default:
		return ""
	}
}

func renderUserMessageCompact(msg session.Message) string {
	promptStyle := lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true)

	textStyle := lipgloss.NewStyle().
		Foreground(colorText)

	text := msg.TextContent()
	if text == "" {
		return ""
	}

	// Truncate long messages
	if len(text) > 200 {
		text = text[:197] + "..."
	}

	prompt := promptStyle.Render("> ")
	content := textStyle.Render(text)

	return prompt + content
}

func renderAssistantMessageCompact(msg session.Message, agentHandle string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	textStyle := lipgloss.NewStyle().
		Foreground(colorText)

	header := headerStyle.Render(IconArrowRight + " " + agentHandle)

	// Get text content only, skip reasoning/tools
	text := msg.TextContent()
	if text == "" {
		return header + "\n" + textStyle.Render("(tool use)")
	}

	// Truncate long messages
	if len(text) > 300 {
		text = text[:297] + "..."
	}

	return header + "\n" + textStyle.Render(text)
}

func renderUserMessage(msg session.Message) string {
	promptStyle := lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true)

	textStyle := lipgloss.NewStyle().
		Foreground(colorText)

	text := msg.TextContent()
	if text == "" {
		return ""
	}

	prompt := promptStyle.Render("> ")
	content := textStyle.Render(text)

	return prompt + content
}

func renderAssistantMessage(msg session.Message, agentHandle string) string {
	var parts []string

	// Header with agent handle
	headerStyle := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	header := headerStyle.Render(IconArrowRight + " " + agentHandle)
	parts = append(parts, header)

	// Render each content part
	for _, part := range msg.Parts {
		switch p := part.(type) {
		case session.TextContent:
			if strings.TrimSpace(p.Text) != "" {
				textStyle := lipgloss.NewStyle().Foreground(colorText)
				parts = append(parts, textStyle.Render(p.Text))
			}

		case session.ReasoningContent:
			if strings.TrimSpace(p.Text) != "" {
				parts = append(parts, renderReasoning(p.Text))
			}

		case session.ToolCall:
			parts = append(parts, renderToolCall(p))
		}
	}

	return strings.Join(parts, "\n")
}

func renderToolMessage(msg session.Message) string {
	var parts []string

	for _, part := range msg.Parts {
		if tr, ok := part.(session.ToolResult); ok {
			parts = append(parts, renderToolResult(tr))
		}
	}

	return strings.Join(parts, "\n")
}

func renderReasoning(text string) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(colorSecondary).
		Italic(true)

	contentStyle := lipgloss.NewStyle().
		Foreground(colorTextDim).
		Italic(true)

	label := labelStyle.Render(IconThinking + " Thinking:")

	// Truncate long reasoning
	content := text
	lines := strings.Split(content, "\n")
	if len(lines) > 5 {
		content = strings.Join(lines[:5], "\n")
		content += fmt.Sprintf("\n... (%d more lines)", len(lines)-5)
	}

	return label + "\n" + contentStyle.Render(content)
}

func renderToolCall(tc session.ToolCall) string {
	toolStyle := lipgloss.NewStyle().
		Foreground(colorTertiary).
		Bold(true)

	cmdStyle := lipgloss.NewStyle().
		Foreground(colorMuted)

	var label string
	var detail string

	if tc.Name == "bash" {
		label = toolStyle.Render(IconBash + " bash")
		// Try to extract command from input JSON
		var input struct {
			Command     string `json:"command"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal([]byte(tc.Input), &input); err == nil {
			if input.Description != "" {
				label += " " + cmdStyle.Render("· "+input.Description)
			}
			if input.Command != "" {
				cmd := input.Command
				if len(cmd) > 60 {
					cmd = cmd[:57] + "..."
				}
				detail = cmdStyle.Render("  $ " + cmd)
			}
		}
	} else {
		label = toolStyle.Render(IconTool + " " + tc.Name)
	}

	if detail != "" {
		return label + "\n" + detail
	}
	return label
}

func renderToolResult(tr session.ToolResult) string {
	var statusIcon string
	var statusColor lipgloss.Color

	if tr.IsError {
		statusIcon = IconError
		statusColor = colorError
	} else {
		statusIcon = IconSuccess
		statusColor = colorSuccess
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	outputStyle := lipgloss.NewStyle().Foreground(colorTextDim)

	status := statusStyle.Render("  " + statusIcon)

	// Show truncated output
	output := tr.Content
	if output != "" {
		lines := strings.Split(output, "\n")
		if len(lines) > 3 {
			output = strings.Join(lines[:3], "\n")
			output += fmt.Sprintf("\n  ... (%d more lines)", len(lines)-3)
		}
		// Indent output
		indented := ""
		for _, line := range strings.Split(output, "\n") {
			if len(line) > 80 {
				line = line[:77] + "..."
			}
			indented += "  " + line + "\n"
		}
		output = strings.TrimSuffix(indented, "\n")

		return status + "\n" + outputStyle.Render(output)
	}

	return status
}
