package run

import (
	"encoding/json"
	"strings"
	"time"

	"charm.land/fantasy"
)

// FantasyAdapter adapts Fantasy's callback-based streaming to the StreamWriter interface.
// It receives Fantasy callbacks and translates them to StreamWriter calls.
type FantasyAdapter struct {
	writer           StreamWriter
	reasoningStart   time.Time
	reasoningContent strings.Builder
}

// NewFantasyAdapter creates an adapter that forwards Fantasy callbacks to the writer.
func NewFantasyAdapter(writer StreamWriter) *FantasyAdapter {
	return &FantasyAdapter{writer: writer}
}

// OnTextDelta is called by Fantasy for each text chunk.
func (a *FantasyAdapter) OnTextDelta(id, text string) error {
	a.writer.WriteText(text)
	return nil
}

// OnTextEnd is called by Fantasy when text streaming ends.
func (a *FantasyAdapter) OnTextEnd(id string) error {
	a.writer.WriteTextDone("")
	return nil
}

// OnReasoningStart is called by Fantasy when reasoning/thinking begins.
func (a *FantasyAdapter) OnReasoningStart(id string) error {
	a.reasoningStart = time.Now()
	a.reasoningContent.Reset()
	return nil
}

// OnReasoningDelta is called by Fantasy for each reasoning chunk.
func (a *FantasyAdapter) OnReasoningDelta(id, text string) error {
	a.reasoningContent.WriteString(text)
	a.writer.WriteReasoning(text)
	return nil
}

// OnReasoningEnd is called by Fantasy when reasoning/thinking ends.
func (a *FantasyAdapter) OnReasoningEnd(id string, duration time.Duration) error {
	// Use provided duration if available, otherwise calculate from start time
	if duration == 0 {
		duration = time.Since(a.reasoningStart)
	}
	a.writer.WriteReasoningDone(a.reasoningContent.String(), duration)
	return nil
}

// OnToolCall is called by Fantasy when a tool is invoked.
func (a *FantasyAdapter) OnToolCall(tc fantasy.ToolCallContent) error {
	call := ToolCall{
		ID:    tc.ToolCallID,
		Name:  tc.ToolName,
		Input: tc.Input,
	}

	// Extract bash-specific params
	if tc.ToolName == "bash" {
		var params struct {
			Command     string `json:"command"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal([]byte(tc.Input), &params); err == nil {
			call.Command = params.Command
			call.Description = params.Description
		}
	}

	a.writer.WriteToolStart(call)
	return nil
}

// OnToolResult is called by Fantasy when a tool completes.
func (a *FantasyAdapter) OnToolResult(result fantasy.ToolResultContent, duration time.Duration) error {
	output := ""
	errStr := ""
	isError := result.Result != nil && result.Result.GetType() == fantasy.ToolResultContentTypeError

	if result.Result != nil {
		switch text := result.Result.(type) {
		case *fantasy.ToolResultOutputContentText:
			output = text.Text
		case fantasy.ToolResultOutputContentText:
			output = text.Text
		}
	}

	if isError {
		errStr = output
	}

	tr := ToolResult{
		ID:       result.ToolCallID,
		Name:     result.ToolName,
		Output:   output,
		Error:    errStr,
		Duration: duration,
		Metadata: result.ClientMetadata,
	}

	a.writer.WriteToolResult(tr)
	return nil
}

// OnAgentStart is called by Fantasy when a sub-agent is invoked.
func (a *FantasyAdapter) OnAgentStart(handle, prompt string) error {
	a.writer.WriteAgentStart(handle, prompt)
	return nil
}

// OnAgentEnd is called by Fantasy when a sub-agent completes.
func (a *FantasyAdapter) OnAgentEnd(handle string, duration time.Duration, err error) error {
	a.writer.WriteAgentEnd(handle, duration, err)
	return nil
}

// OnMemoryEvent is called when a memory event occurs.
func (a *FantasyAdapter) OnMemoryEvent(event string, count int) error {
	a.writer.WriteMemoryEvent(event, count)
	return nil
}

// OnError is called when an error occurs.
func (a *FantasyAdapter) OnError(err error) error {
	a.writer.WriteError(err)
	return nil
}

// Verify FantasyAdapter implements StreamHandler (for backward compatibility during transition)
var _ StreamHandler = (*FantasyAdapter)(nil)
