package memory

import (
	"testing"
)

func TestNewStoreMemoryTool(t *testing.T) {
	// Test that tool can be created (nil service is allowed)
	tool := NewStoreMemoryTool(ToolConfig{
		Service:     nil,
		AgentHandle: "@test",
		PathScope:   "/workspace",
		SessionID:   "test-session",
	})

	if tool == nil {
		t.Fatal("expected tool to be created")
	}
}

func TestNewSearchMemoryTool(t *testing.T) {
	// Test that tool can be created
	tool := NewSearchMemoryTool(ToolConfig{
		Service:     nil,
		AgentHandle: "@test",
		PathScope:   "/workspace",
		SessionID:   "test-session",
	})

	if tool == nil {
		t.Fatal("expected tool to be created")
	}
}

func TestStoreResult_String(t *testing.T) {
	result := StoreResult{
		ID:      "abc123",
		Message: "Stored memory [fact]: User prefers Go over Python",
	}

	s := result.String()
	if s != result.Message {
		t.Errorf("expected %q, got %q", result.Message, s)
	}
}

func TestSearchResult_String(t *testing.T) {
	t.Run("empty results", func(t *testing.T) {
		result := SearchResult{
			Memories: nil,
			Message:  "Found 0 matching memories",
		}

		s := result.String()
		if s != "No matching memories found." {
			t.Errorf("expected empty message, got %q", s)
		}
	})

	t.Run("with results", func(t *testing.T) {
		result := SearchResult{
			Memories: []MemoryMatch{
				{ID: "1", Content: "User prefers tabs", Category: "preference", Similarity: 0.95},
				{ID: "2", Content: "Project uses Go", Category: "fact", Similarity: 0.85},
			},
			Message: "Found 2 matching memories",
		}

		s := result.String()
		if s == "" || s == "No matching memories found." {
			t.Errorf("expected results string, got %q", s)
		}
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exactly10!", 10, "exactly10!"},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.expected)
		}
	}
}
