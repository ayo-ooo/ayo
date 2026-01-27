package chat

import tea "github.com/charmbracelet/bubbletea"

// Messages for communication between the chat TUI and external components.

// ToolCallStartMsg indicates a tool is starting execution.
type ToolCallStartMsg struct {
	Name        string
	Description string
	Command     string // For bash tools
}

// ToolCallResultMsg indicates a tool has completed.
type ToolCallResultMsg struct {
	Name     string
	Output   string
	Error    string
	Duration string
}

// ReasoningStartMsg indicates reasoning/thinking has started.
type ReasoningStartMsg struct{}

// ReasoningDeltaMsg contains a chunk of reasoning content.
type ReasoningDeltaMsg struct {
	Delta string
}

// ReasoningEndMsg indicates reasoning has completed.
type ReasoningEndMsg struct {
	Duration string
}

// TextDeltaMsg contains a chunk of response text.
type TextDeltaMsg struct {
	Delta string
}

// TextEndMsg indicates text streaming has completed.
type TextEndMsg struct{}

// ErrorMsg indicates an error occurred.
type ErrorMsg struct {
	Error error
}

// SubAgentStartMsg indicates a sub-agent is being invoked.
type SubAgentStartMsg struct {
	Handle string
	Prompt string
}

// SubAgentEndMsg indicates a sub-agent has completed.
type SubAgentEndMsg struct {
	Handle   string
	Duration string
	Error    bool
}

// MemoryEventMsg indicates a memory operation.
type MemoryEventMsg struct {
	Type string // "created", "skipped", "superseded", "failed"
}

// Cmd helpers for sending messages to the TUI from callbacks.

// SendToolCallStart creates a command to signal tool start.
func SendToolCallStart(name, description, command string) tea.Cmd {
	return func() tea.Msg {
		return ToolCallStartMsg{
			Name:        name,
			Description: description,
			Command:     command,
		}
	}
}

// SendToolCallResult creates a command to signal tool completion.
func SendToolCallResult(name, output, errStr, duration string) tea.Cmd {
	return func() tea.Msg {
		return ToolCallResultMsg{
			Name:     name,
			Output:   output,
			Error:    errStr,
			Duration: duration,
		}
	}
}

// SendTextDelta creates a command to send streaming text.
func SendTextDelta(delta string) tea.Cmd {
	return func() tea.Msg {
		return TextDeltaMsg{Delta: delta}
	}
}

// SendTextEnd creates a command to signal text completion.
func SendTextEnd() tea.Cmd {
	return func() tea.Msg {
		return TextEndMsg{}
	}
}
