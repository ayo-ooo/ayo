package pubsub

import (
	"time"
)

// MessageEvent represents events related to chat messages.
type MessageEvent struct {
	// MessageID is the unique identifier for the message.
	MessageID string
	// SessionID is the session this message belongs to.
	SessionID string
	// Role is the message role (user, assistant, system).
	Role string
	// Content is the message content (may be partial for streaming).
	Content string
	// Parts contains structured message parts as interface{} to avoid import cycles.
	Parts []interface{}
	// ParentMessageID is set for nested agent messages.
	ParentMessageID string
	// ToolCallID is set if this message is related to a tool call.
	ToolCallID string
}

// ToolEvent represents events related to tool calls.
type ToolEvent struct {
	// ToolCallID is the unique identifier for the tool call.
	ToolCallID string
	// Name is the tool name (e.g., "bash", "todo").
	Name string
	// Input is the JSON input to the tool.
	Input string
	// Output is the tool output (set on completion).
	Output string
	// Error is set if the tool call failed.
	Error string
	// IsError indicates if the tool returned an error.
	IsError bool
	// Metadata is additional JSON metadata from the tool.
	Metadata string
	// Duration is the execution time (set on completion).
	Duration time.Duration
	// Description is a human-readable description of what the tool is doing.
	Description string
	// Command is the command being executed (for bash).
	Command string
	// ParentMessageID links this tool call to its parent message.
	ParentMessageID string
	// ParentToolCallID is set for nested tool calls.
	ParentToolCallID string
	// Cancelled indicates the tool call was cancelled.
	Cancelled bool
	// PermissionRequested indicates permission was requested.
	PermissionRequested bool
	// PermissionGranted indicates permission was granted.
	PermissionGranted bool
}

// ReasoningEvent represents events related to extended thinking/reasoning.
type ReasoningEvent struct {
	// ID is the unique identifier for this reasoning block.
	ID string
	// Content is the reasoning content (may be partial for streaming).
	Content string
	// Duration is the total reasoning time (set on completion).
	Duration time.Duration
}

// MemoryEvent represents events related to memory operations.
type MemoryEvent struct {
	// MemoryID is the unique identifier for the memory.
	MemoryID string
	// Content is the memory content.
	Content string
	// Category is the memory category (preference, fact, correction, pattern).
	Category string
	// Scope is the memory scope (global, agent, path).
	Scope string
	// Operation describes what happened (created, skipped, superseded, failed).
	Operation string
	// Count is used for batch operations (e.g., "5 memories recalled").
	Count int
}

// AgentEvent represents events related to sub-agent calls.
type AgentEvent struct {
	// AgentHandle is the handle of the agent being called.
	AgentHandle string
	// Prompt is the prompt sent to the agent.
	Prompt string
	// SessionID is the sub-agent's session ID.
	SessionID string
	// ParentSessionID is the parent session that spawned this agent.
	ParentSessionID string
	// ParentToolCallID links this agent call to a tool call.
	ParentToolCallID string
	// Duration is the total execution time (set on completion).
	Duration time.Duration
	// Error is set if the agent call failed.
	Error string
}

// TextDeltaEvent represents streaming text content.
type TextDeltaEvent struct {
	// ID identifies the text stream.
	ID string
	// Delta is the text chunk.
	Delta string
	// Final indicates this is the last chunk.
	Final bool
}
