package run

import (
	"errors"
	"testing"
	"time"

	"charm.land/fantasy"
)

func TestNullStreamHandler_Implements(t *testing.T) {
	// Verify NullStreamHandler implements StreamHandler
	var _ StreamHandler = NullStreamHandler{}
}

func TestNullStreamHandler_AllMethodsReturnNil(t *testing.T) {
	h := NullStreamHandler{}

	tests := []struct {
		name string
		fn   func() error
	}{
		{"OnTextDelta", func() error { return h.OnTextDelta("id", "text") }},
		{"OnTextEnd", func() error { return h.OnTextEnd("id") }},
		{"OnReasoningStart", func() error { return h.OnReasoningStart("id") }},
		{"OnReasoningDelta", func() error { return h.OnReasoningDelta("id", "text") }},
		{"OnReasoningEnd", func() error { return h.OnReasoningEnd("id", time.Second) }},
		{"OnToolCall", func() error { return h.OnToolCall(fantasy.ToolCallContent{}) }},
		{"OnToolResult", func() error { return h.OnToolResult(fantasy.ToolResultContent{}, time.Second) }},
		{"OnAgentStart", func() error { return h.OnAgentStart("@ayo", "prompt") }},
		{"OnAgentEnd", func() error { return h.OnAgentEnd("@ayo", time.Second, nil) }},
		{"OnAgentEnd with error", func() error { return h.OnAgentEnd("@ayo", time.Second, errors.New("test")) }},
		{"OnMemoryEvent", func() error { return h.OnMemoryEvent("stored", 1) }},
		{"OnError", func() error { return h.OnError(errors.New("test")) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil {
				t.Errorf("%s() = %v, want nil", tt.name, err)
			}
		})
	}
}

func TestToolCallInfo(t *testing.T) {
	info := ToolCallInfo{
		ID:          "call-123",
		Name:        "bash",
		Description: "Running tests",
		Command:     "go test ./...",
		Input:       `{"command": "go test ./..."}`,
	}

	if info.ID != "call-123" {
		t.Errorf("ID = %q, want %q", info.ID, "call-123")
	}
	if info.Name != "bash" {
		t.Errorf("Name = %q, want %q", info.Name, "bash")
	}
	if info.Description != "Running tests" {
		t.Errorf("Description = %q, want %q", info.Description, "Running tests")
	}
}
