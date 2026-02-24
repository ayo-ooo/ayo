package interactive

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ToolState represents the execution state of a tool call.
type ToolState int

const (
	ToolStateRunning ToolState = iota
	ToolStateSuccess
	ToolStateError
)

// ToolDisplay handles the inline display of tool calls.
type ToolDisplay struct {
	name    string
	input   string
	state   ToolState
	result  string
	depth   int // For nested tool calls
}

// NewToolDisplay creates a new tool display.
func NewToolDisplay(name, input string) *ToolDisplay {
	return &ToolDisplay{
		name:  name,
		input: input,
		state: ToolStateRunning,
	}
}

// SetResult sets the tool result and state.
func (t *ToolDisplay) SetResult(result string, success bool) {
	t.result = result
	if success {
		t.state = ToolStateSuccess
	} else {
		t.state = ToolStateError
	}
}

// View renders the tool display.
func (t *ToolDisplay) View() string {
	var b strings.Builder

	indent := strings.Repeat("  ", t.depth+1)
	childIndent := strings.Repeat("  ", t.depth+2)

	// Style definitions
	toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#67e8f9"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))

	// Tool line
	toolName := toolStyle.Render(t.name)
	inputSummary := summarizeInput(t.name, t.input)

	switch t.state {
	case ToolStateRunning:
		b.WriteString(fmt.Sprintf("%s⋯ %s: %s\n", indent, toolName, inputSummary))
	default:
		b.WriteString(fmt.Sprintf("%s▸ %s: %s\n", indent, toolName, inputSummary))
	}

	// Result line (if complete)
	if t.state == ToolStateSuccess {
		resultSummary := summarizeResult(t.name, t.result)
		b.WriteString(fmt.Sprintf("%s└─ %s\n", childIndent, successStyle.Render(resultSummary)))
	} else if t.state == ToolStateError {
		b.WriteString(fmt.Sprintf("%s└─ ✗ %s\n", childIndent, errorStyle.Render(t.result)))
	}

	_ = dimStyle // Keeping for future use
	return b.String()
}

// ToolSummarizer summarizes tool output to a short string.
type ToolSummarizer func(output string) string

// summarizers provides tool-specific output summarization.
var summarizers = map[string]ToolSummarizer{
	"bash": func(output string) string {
		lines := strings.Count(output, "\n")
		if lines > 5 {
			return fmt.Sprintf("%d lines of output", lines)
		}
		// Return first non-empty line, truncated
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				if len(line) > 60 {
					return line[:57] + "..."
				}
				return line
			}
		}
		return "completed"
	},
	"view": func(output string) string {
		lines := strings.Count(output, "\n")
		return fmt.Sprintf("Read %d lines", lines)
	},
	"edit": func(output string) string {
		return "File updated"
	},
	"multiedit": func(output string) string {
		return "Multiple edits applied"
	},
	"write": func(output string) string {
		return "File written"
	},
	"grep": func(output string) string {
		lines := strings.Count(output, "\n")
		if lines == 0 {
			return "No matches"
		}
		return fmt.Sprintf("%d matches", lines)
	},
	"glob": func(output string) string {
		lines := strings.Count(output, "\n")
		return fmt.Sprintf("%d files", lines)
	},
	"ls": func(output string) string {
		lines := strings.Count(output, "\n")
		return fmt.Sprintf("%d items", lines)
	},
	"delegate": func(output string) string {
		if strings.Contains(output, "error") || strings.Contains(output, "Error") {
			return "Failed"
		}
		return "Completed"
	},
	"todos": func(output string) string {
		return "Updated"
	},
	"memory": func(output string) string {
		return "Done"
	},
}

// summarizeResult summarizes tool output based on tool type.
func summarizeResult(toolName, output string) string {
	if summarizer, ok := summarizers[toolName]; ok {
		return summarizer(output)
	}

	// Default summarizer
	output = strings.TrimSpace(output)
	if output == "" {
		return "completed"
	}

	// First line, truncated
	lines := strings.Split(output, "\n")
	first := strings.TrimSpace(lines[0])
	if len(first) > 60 {
		return first[:57] + "..."
	}
	return first
}

// inputSummarizers provides tool-specific input summarization.
var inputSummarizers = map[string]ToolSummarizer{
	"bash": func(input string) string {
		// Extract command, truncate if long
		if len(input) > 50 {
			return input[:47] + "..."
		}
		return input
	},
	"view": func(input string) string {
		return input
	},
	"edit": func(input string) string {
		// Just show the file path
		return input
	},
	"grep": func(input string) string {
		return input
	},
	"delegate": func(input string) string {
		return input
	},
}

// summarizeInput summarizes tool input for display.
func summarizeInput(toolName, input string) string {
	input = strings.TrimSpace(input)
	if summarizer, ok := inputSummarizers[toolName]; ok {
		return summarizer(input)
	}

	// Default: truncate
	if len(input) > 50 {
		return input[:47] + "..."
	}
	return input
}

// FormatToolCall formats a tool call for display.
func FormatToolCall(name, input string, result string, success bool) string {
	display := NewToolDisplay(name, input)
	display.SetResult(result, success)
	return display.View()
}

// FormatToolCallRunning formats a running tool call.
func FormatToolCallRunning(name, input string) string {
	display := NewToolDisplay(name, input)
	return display.View()
}
