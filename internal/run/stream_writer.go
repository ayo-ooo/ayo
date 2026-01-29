package run

import (
	"time"
)

// StreamWriter is the unified interface for streaming output from agent execution.
// Both TUI mode (ChannelWriter) and non-interactive mode (PrintWriter) implement this.
// This replaces the callback-based StreamHandler with a simpler push-based API.
type StreamWriter interface {
	// Text streaming
	WriteText(delta string)
	WriteTextDone(content string)

	// Reasoning/thinking (for models that support extended thinking)
	WriteReasoning(delta string)
	WriteReasoningDone(content string, duration time.Duration)

	// Tool calls
	WriteToolStart(call ToolCall)
	WriteToolResult(result ToolResult)

	// Sub-agent calls
	WriteAgentStart(handle, prompt string)
	WriteAgentEnd(handle string, duration time.Duration, err error)

	// Memory events
	WriteMemoryEvent(event string, count int)

	// Errors and completion
	WriteError(err error)
	WriteDone(response string)
}

// ToolCall represents a tool invocation in progress.
type ToolCall struct {
	ID          string
	Name        string
	Description string
	Command     string // For bash tool
	Input       string
	ParentID    string // For nested calls (sub-agent tool calls)
}

// ToolResult represents a completed tool invocation.
type ToolResult struct {
	ID       string
	Name     string
	Output   string
	Error    string
	Duration time.Duration
	Metadata string // Client metadata (e.g., todo list state)
}

// NullWriter is a no-op writer for testing or silent mode.
type NullWriter struct{}

func (NullWriter) WriteText(delta string)                                       {}
func (NullWriter) WriteTextDone(content string)                                 {}
func (NullWriter) WriteReasoning(delta string)                                  {}
func (NullWriter) WriteReasoningDone(content string, duration time.Duration)    {}
func (NullWriter) WriteToolStart(call ToolCall)                                 {}
func (NullWriter) WriteToolResult(result ToolResult)                            {}
func (NullWriter) WriteAgentStart(handle, prompt string)                        {}
func (NullWriter) WriteAgentEnd(handle string, duration time.Duration, err error) {}
func (NullWriter) WriteMemoryEvent(event string, count int)                     {}
func (NullWriter) WriteError(err error)                                         {}
func (NullWriter) WriteDone(response string)                                    {}

// Verify NullWriter implements StreamWriter
var _ StreamWriter = NullWriter{}
