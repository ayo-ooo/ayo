// Package jsonl provides JSONL-based session file storage.
// Each session is stored as a JSONL file with a header line containing session
// metadata followed by message lines.
package jsonl

import (
	"encoding/json"
	"errors"
	"time"
)

// Common errors
var (
	ErrEmptyFile       = errors.New("empty session file")
	ErrInvalidHeader   = errors.New("invalid session header")
	ErrSessionNotFound = errors.New("session not found")
)

// LineType identifies the type of line in a JSONL session file.
type LineType string

const (
	LineTypeSession LineType = "session"
	LineTypeMessage LineType = "message"
)

// SessionHeader represents the first line of a session JSONL file.
type SessionHeader struct {
	Type             LineType   `json:"type"`
	ID               string     `json:"id"`
	AgentHandle      string     `json:"agent_handle"`
	Title            string     `json:"title"`
	Source           string     `json:"source"` // "ayo", "crush", "crush-via-ayo"
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
	ChainDepth       int        `json:"chain_depth"`
	ChainSource      string     `json:"chain_source,omitempty"`
	InputSchema      *string    `json:"input_schema,omitempty"`
	OutputSchema     *string    `json:"output_schema,omitempty"`
	StructuredInput  *string    `json:"structured_input,omitempty"`
	StructuredOutput *string    `json:"structured_output,omitempty"`
	MessageCount     int        `json:"message_count"`
}

// MessageLine represents a message line in a session JSONL file.
type MessageLine struct {
	Type       LineType        `json:"type"`
	ID         string          `json:"id"`
	Role       string          `json:"role"` // "system", "user", "assistant", "tool"
	Model      string          `json:"model,omitempty"`
	Provider   string          `json:"provider,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	FinishedAt *time.Time      `json:"finished_at,omitempty"`
	Parts      json.RawMessage `json:"parts"` // Preserve exact JSON structure
}

// ContentPart represents a part of a message.
type ContentPart struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// TextData represents text content data.
type TextData struct {
	Text string `json:"text"`
}

// ReasoningData represents reasoning/thinking content.
type ReasoningData struct {
	Text       string     `json:"text"`
	Signature  string     `json:"signature,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// ToolCallData represents a tool call.
type ToolCallData struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Input            json.RawMessage `json:"input"`
	ProviderExecuted bool            `json:"provider_executed,omitempty"`
	Finished         bool            `json:"finished,omitempty"`
	StartedAt        *time.Time      `json:"started_at,omitempty"`
	FinishedAt       *time.Time      `json:"finished_at,omitempty"`
}

// ToolResultData represents a tool result.
type ToolResultData struct {
	ToolCallID string          `json:"tool_call_id"`
	Name       string          `json:"name"`
	Content    json.RawMessage `json:"content"`
	IsError    bool            `json:"is_error,omitempty"`
}

// FileData represents file content.
type FileData struct {
	Filename  string `json:"filename"`
	Data      string `json:"data"` // base64 encoded
	MediaType string `json:"media_type"`
}

// FinishData represents a finish marker.
type FinishData struct {
	Reason  string     `json:"reason"`
	Time    *time.Time `json:"time,omitempty"`
	Message *string    `json:"message,omitempty"`
}

// ParseLine parses a JSONL line and returns the type and raw data.
func ParseLine(line []byte) (LineType, json.RawMessage, error) {
	var base struct {
		Type LineType `json:"type"`
	}
	if err := json.Unmarshal(line, &base); err != nil {
		return "", nil, err
	}
	return base.Type, line, nil
}

// ParseSessionHeader parses a session header from a JSONL line.
func ParseSessionHeader(line []byte) (*SessionHeader, error) {
	var header SessionHeader
	if err := json.Unmarshal(line, &header); err != nil {
		return nil, err
	}
	if header.Type != LineTypeSession {
		return nil, ErrInvalidHeader
	}
	return &header, nil
}

// ParseMessageLine parses a message from a JSONL line.
func ParseMessageLine(line []byte) (*MessageLine, error) {
	var msg MessageLine
	if err := json.Unmarshal(line, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// MarshalLine marshals any line type to JSON with no trailing newline.
func MarshalLine(v any) ([]byte, error) {
	return json.Marshal(v)
}
