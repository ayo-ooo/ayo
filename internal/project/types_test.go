package project

import (
	"testing"
)

func TestValidationError_Error_WithLine(t *testing.T) {
	e := &ValidationError{
		File:    "config.toml",
		Message: "invalid syntax",
		Line:    10,
	}

	got := e.Error()

	if got == "" {
		t.Error("Error() returned empty string")
	}

	if len(got) < len("config.toml") {
		t.Errorf("Error() should contain filename, got: %q", got)
	}
}

func TestValidationError_Error_WithoutLine(t *testing.T) {
	e := &ValidationError{
		File:    "system.md",
		Message: "file is empty",
		Line:    0,
	}

	got := e.Error()

	if got == "" {
		t.Error("Error() returned empty string")
	}

	if len(got) < len("system.md") {
		t.Errorf("Error() should contain filename, got: %q", got)
	}
}

func TestSchema_UnmarshalJSON_Valid(t *testing.T) {
	jsonData := `{"type": "object", "properties": {"name": {"type": "string"}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if s.Content == nil {
		t.Error("Content should not be nil")
	}

	if s.Parsed == nil {
		t.Error("Parsed should not be nil")
	}
}

func TestSchema_UnmarshalJSON_Invalid(t *testing.T) {
	jsonData := `{invalid}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	if err == nil {
		t.Error("UnmarshalJSON() expected error for invalid JSON")
	}
}

func TestHookType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		hookType HookType
		want     string
	}{
		{"agent-start", HookAgentStart, "agent-start"},
		{"agent-finish", HookAgentFinish, "agent-finish"},
		{"agent-error", HookAgentError, "agent-error"},
		{"text-start", HookTextStart, "text-start"},
		{"text-delta", HookTextDelta, "text-delta"},
		{"text-end", HookTextEnd, "text-end"},
		{"tool-call", HookToolCall, "tool-call"},
		{"tool-result", HookToolResult, "tool-result"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.hookType) != tt.want {
				t.Errorf("HookType = %q, want %q", tt.hookType, tt.want)
			}
		})
	}
}
