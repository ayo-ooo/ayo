package run

import (
	"time"

	"charm.land/fantasy"
)

// StreamHandler abstracts the output handling for agent execution.
// This allows switching between TUI mode (publishing events) and
// print mode (writing directly to stdout).
type StreamHandler interface {
	// Text streaming
	OnTextDelta(id, text string) error
	OnTextEnd(id string) error

	// Reasoning/thinking
	OnReasoningStart(id string) error
	OnReasoningDelta(id, text string) error
	OnReasoningEnd(id string, duration time.Duration) error

	// Tool calls
	OnToolCall(tc fantasy.ToolCallContent) error
	OnToolResult(result fantasy.ToolResultContent, duration time.Duration) error

	// Sub-agent calls
	OnAgentStart(handle, prompt string) error
	OnAgentEnd(handle string, duration time.Duration, err error) error

	// Memory events
	OnMemoryEvent(event string, count int) error

	// Errors
	OnError(err error) error
}

// ToolCallInfo contains display information for a tool call.
type ToolCallInfo struct {
	ID          string
	Name        string
	Description string
	Command     string
	Input       string
}

// NullStreamHandler is a no-op handler for testing or silent mode.
type NullStreamHandler struct{}

func (NullStreamHandler) OnTextDelta(id, text string) error                        { return nil }
func (NullStreamHandler) OnTextEnd(id string) error                                { return nil }
func (NullStreamHandler) OnReasoningStart(id string) error                         { return nil }
func (NullStreamHandler) OnReasoningDelta(id, text string) error                   { return nil }
func (NullStreamHandler) OnReasoningEnd(id string, duration time.Duration) error   { return nil }
func (NullStreamHandler) OnToolCall(tc fantasy.ToolCallContent) error              { return nil }
func (NullStreamHandler) OnToolResult(result fantasy.ToolResultContent, duration time.Duration) error {
	return nil
}
func (NullStreamHandler) OnAgentStart(handle, prompt string) error                     { return nil }
func (NullStreamHandler) OnAgentEnd(handle string, duration time.Duration, err error) error {
	return nil
}
func (NullStreamHandler) OnMemoryEvent(event string, count int) error { return nil }
func (NullStreamHandler) OnError(err error) error                     { return nil }
