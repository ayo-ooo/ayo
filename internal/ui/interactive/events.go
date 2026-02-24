// Package interactive provides event rendering for interactive mode.
package interactive

import (
	"time"
)

// EventType represents the type of event from agent interactions.
type EventType string

const (
	EventText         EventType = "text"          // Streaming text
	EventToolStart    EventType = "tool_start"    // Tool invocation starting
	EventToolProgress EventType = "tool_progress" // Tool progress update
	EventToolComplete EventType = "tool_complete" // Tool finished
	EventToolError    EventType = "tool_error"    // Tool failed
	EventThinking     EventType = "thinking"      // Agent thinking indicator
	EventDelegate     EventType = "delegate"      // Delegation to another agent
	EventPlanUpdate   EventType = "plan_update"   // Todo/ticket change
	EventMemory       EventType = "memory"        // Memory operation
	EventError        EventType = "error"         // Error occurred
	EventComplete     EventType = "complete"      // Response complete
)

// Event represents a single event from agent interactions.
type Event struct {
	Type      EventType
	Timestamp time.Time
	Data      any
}

// TextData contains streaming text content.
type TextData struct {
	Text string
}

// ToolStartData contains information about a starting tool invocation.
type ToolStartData struct {
	ID      string // Unique tool call ID
	Name    string // Tool name (e.g., "grep", "view", "edit")
	Summary string // One-line summary of what the tool is doing
	Params  any    // Optional: raw parameters
}

// ToolProgressData contains progress update for a running tool.
type ToolProgressData struct {
	ID       string // Tool call ID
	Message  string // Progress message
	Progress float64 // Optional: progress percentage (0-1)
}

// ToolCompleteData contains the result of a completed tool invocation.
type ToolCompleteData struct {
	ID      string // Tool call ID
	Name    string // Tool name
	Summary string // One-line result summary
	Success bool   // Whether the tool succeeded
	Output  string // Optional: full output (may be truncated for display)
}

// ToolErrorData contains error information for a failed tool invocation.
type ToolErrorData struct {
	ID      string // Tool call ID
	Name    string // Tool name
	Error   string // Error message
	Details string // Optional: additional details
}

// ThinkingData contains information about agent thinking/reasoning.
type ThinkingData struct {
	Message string
	Active  bool // Whether currently thinking
}

// DelegateData contains information about delegation to another agent.
type DelegateData struct {
	Agent   string // Target agent name
	Task    string // What was delegated
	Started bool   // True if starting, false if completed
	Result  string // Result if completed
}

// PlanUpdateData contains planning/todo updates.
type PlanUpdateData struct {
	Items []PlanItem
}

// PlanItem represents a single item in a plan.
type PlanItem struct {
	ID       string
	Content  string
	Status   PlanItemStatus
	Progress string // Optional: progress indicator
}

// PlanItemStatus represents the status of a plan item.
type PlanItemStatus string

const (
	PlanItemPending    PlanItemStatus = "pending"
	PlanItemInProgress PlanItemStatus = "in_progress"
	PlanItemCompleted  PlanItemStatus = "completed"
)

// MemoryData contains information about memory operations.
type MemoryData struct {
	Operation string // "read", "write", "delete"
	Path      string // Memory path
	Message   string // Description
}

// ErrorData contains error information.
type ErrorData struct {
	Message string
	Details string
}

// CompleteData contains information about response completion.
type CompleteData struct {
	TokensUsed int
	Duration   time.Duration
}

// NewEvent creates a new event with the current timestamp.
func NewEvent(eventType EventType, data any) Event {
	return Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewTextEvent creates a text streaming event.
func NewTextEvent(text string) Event {
	return NewEvent(EventText, &TextData{Text: text})
}

// NewToolStartEvent creates a tool start event.
func NewToolStartEvent(id, name, summary string) Event {
	return NewEvent(EventToolStart, &ToolStartData{
		ID:      id,
		Name:    name,
		Summary: summary,
	})
}

// NewToolCompleteEvent creates a tool completion event.
func NewToolCompleteEvent(id, name, summary string, success bool) Event {
	return NewEvent(EventToolComplete, &ToolCompleteData{
		ID:      id,
		Name:    name,
		Summary: summary,
		Success: success,
	})
}

// NewToolErrorEvent creates a tool error event.
func NewToolErrorEvent(id, name, errMsg string) Event {
	return NewEvent(EventToolError, &ToolErrorData{
		ID:    id,
		Name:  name,
		Error: errMsg,
	})
}

// NewErrorEvent creates an error event.
func NewErrorEvent(message, details string) Event {
	return NewEvent(EventError, &ErrorData{
		Message: message,
		Details: details,
	})
}

// NewCompleteEvent creates a completion event.
func NewCompleteEvent(tokensUsed int, duration time.Duration) Event {
	return NewEvent(EventComplete, &CompleteData{
		TokensUsed: tokensUsed,
		Duration:   duration,
	})
}
