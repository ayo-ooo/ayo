package pubsub

import (
	"testing"
	"time"
)

func TestMessageEvent_Fields(t *testing.T) {
	e := MessageEvent{
		MessageID:       "msg-123",
		SessionID:       "session-456",
		Role:            "assistant",
		Content:         "Hello, world!",
		Parts:           []any{"part1", "part2"},
		ParentMessageID: "msg-parent",
		ToolCallID:      "tool-789",
	}

	if e.MessageID != "msg-123" {
		t.Errorf("MessageID = %q, want %q", e.MessageID, "msg-123")
	}
	if e.SessionID != "session-456" {
		t.Errorf("SessionID = %q, want %q", e.SessionID, "session-456")
	}
	if e.Role != "assistant" {
		t.Errorf("Role = %q, want %q", e.Role, "assistant")
	}
	if e.Content != "Hello, world!" {
		t.Errorf("Content = %q, want %q", e.Content, "Hello, world!")
	}
	if len(e.Parts) != 2 {
		t.Errorf("len(Parts) = %d, want 2", len(e.Parts))
	}
}

func TestToolEvent_Fields(t *testing.T) {
	e := ToolEvent{
		ToolCallID:          "call-123",
		Name:                "bash",
		Input:               `{"command": "ls"}`,
		Output:              "file1.go\nfile2.go",
		Error:               "",
		IsError:             false,
		Metadata:            `{"exit_code": 0}`,
		Duration:            500 * time.Millisecond,
		Description:         "Listing files",
		Command:             "ls",
		ParentMessageID:     "msg-456",
		ParentToolCallID:    "",
		Cancelled:           false,
		PermissionRequested: false,
		PermissionGranted:   false,
	}

	if e.ToolCallID != "call-123" {
		t.Errorf("ToolCallID = %q, want %q", e.ToolCallID, "call-123")
	}
	if e.Name != "bash" {
		t.Errorf("Name = %q, want %q", e.Name, "bash")
	}
	if e.Duration != 500*time.Millisecond {
		t.Errorf("Duration = %v, want %v", e.Duration, 500*time.Millisecond)
	}
}

func TestToolEvent_ErrorState(t *testing.T) {
	e := ToolEvent{
		ToolCallID: "call-err",
		Name:       "bash",
		Error:      "command not found",
		IsError:    true,
	}

	if !e.IsError {
		t.Error("IsError should be true")
	}
	if e.Error != "command not found" {
		t.Errorf("Error = %q, want %q", e.Error, "command not found")
	}
}

func TestReasoningEvent_Fields(t *testing.T) {
	e := ReasoningEvent{
		ID:       "reason-123",
		Content:  "Let me think about this...",
		Duration: 2 * time.Second,
	}

	if e.ID != "reason-123" {
		t.Errorf("ID = %q, want %q", e.ID, "reason-123")
	}
	if e.Content != "Let me think about this..." {
		t.Errorf("Content = %q, want %q", e.Content, "Let me think about this...")
	}
	if e.Duration != 2*time.Second {
		t.Errorf("Duration = %v, want %v", e.Duration, 2*time.Second)
	}
}

func TestMemoryEvent_Fields(t *testing.T) {
	e := MemoryEvent{
		MemoryID:  "mem-123",
		Content:   "User prefers dark mode",
		Category:  "preference",
		Scope:     "global",
		Operation: "created",
		Count:     1,
	}

	if e.MemoryID != "mem-123" {
		t.Errorf("MemoryID = %q, want %q", e.MemoryID, "mem-123")
	}
	if e.Category != "preference" {
		t.Errorf("Category = %q, want %q", e.Category, "preference")
	}
	if e.Scope != "global" {
		t.Errorf("Scope = %q, want %q", e.Scope, "global")
	}
	if e.Operation != "created" {
		t.Errorf("Operation = %q, want %q", e.Operation, "created")
	}
}

func TestTextDeltaEvent_Fields(t *testing.T) {
	e := TextDeltaEvent{
		ID:    "text-123",
		Delta: "Hello",
		Final: false,
	}

	if e.ID != "text-123" {
		t.Errorf("ID = %q, want %q", e.ID, "text-123")
	}
	if e.Delta != "Hello" {
		t.Errorf("Delta = %q, want %q", e.Delta, "Hello")
	}
	if e.Final {
		t.Error("Final should be false")
	}
}

func TestTextDeltaEvent_Final(t *testing.T) {
	e := TextDeltaEvent{
		ID:    "text-123",
		Delta: "",
		Final: true,
	}

	if !e.Final {
		t.Error("Final should be true")
	}
}
