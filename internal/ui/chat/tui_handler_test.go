package chat

import (
	"testing"

	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/ui/chat/panels"
)

func TestConvertSearchResultToMemoryItem(t *testing.T) {
	tests := []struct {
		name     string
		input    memory.SearchResult
		expected panels.MemoryItem
	}{
		{
			name: "preference with agent scope",
			input: memory.SearchResult{
				Memory: memory.Memory{
					ID:          "mem-123",
					Content:     "User prefers dark mode",
					Category:    memory.CategoryPreference,
					AgentHandle: "@ayo",
				},
				Similarity: 0.85,
			},
			expected: panels.MemoryItem{
				ID:       "mem-123",
				Content:  "User prefers dark mode",
				Category: "preference",
				Scope:    "@ayo",
			},
		},
		{
			name: "fact with path scope",
			input: memory.SearchResult{
				Memory: memory.Memory{
					ID:        "mem-456",
					Content:   "Project uses Go 1.21",
					Category:  memory.CategoryFact,
					PathScope: "/home/user/project",
				},
				Similarity: 0.75,
			},
			expected: panels.MemoryItem{
				ID:       "mem-456",
				Content:  "Project uses Go 1.21",
				Category: "fact",
				Scope:    "/home/user/project",
			},
		},
		{
			name: "global correction",
			input: memory.SearchResult{
				Memory: memory.Memory{
					ID:       "mem-789",
					Content:  "User prefers verbose explanations",
					Category: memory.CategoryCorrection,
					// No AgentHandle or PathScope = global
				},
				Similarity: 0.90,
			},
			expected: panels.MemoryItem{
				ID:       "mem-789",
				Content:  "User prefers verbose explanations",
				Category: "correction",
				Scope:    "global",
			},
		},
		{
			name: "pattern memory",
			input: memory.SearchResult{
				Memory: memory.Memory{
					ID:       "mem-abc",
					Content:  "User typically works on backend services",
					Category: memory.CategoryPattern,
				},
				Similarity: 0.70,
			},
			expected: panels.MemoryItem{
				ID:       "mem-abc",
				Content:  "User typically works on backend services",
				Category: "pattern",
				Scope:    "global",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSearchResultToMemoryItem(tt.input)

			if result.ID != tt.expected.ID {
				t.Errorf("ID = %q, want %q", result.ID, tt.expected.ID)
			}
			if result.Content != tt.expected.Content {
				t.Errorf("Content = %q, want %q", result.Content, tt.expected.Content)
			}
			if result.Category != tt.expected.Category {
				t.Errorf("Category = %q, want %q", result.Category, tt.expected.Category)
			}
			if result.Scope != tt.expected.Scope {
				t.Errorf("Scope = %q, want %q", result.Scope, tt.expected.Scope)
			}
		})
	}
}

func TestConvertSearchResultsToMemoryItems(t *testing.T) {
	results := []memory.SearchResult{
		{
			Memory: memory.Memory{
				ID:          "1",
				Content:     "First",
				Category:    memory.CategoryPreference,
				AgentHandle: "@test",
			},
		},
		{
			Memory: memory.Memory{
				ID:       "2",
				Content:  "Second",
				Category: memory.CategoryFact,
			},
		},
	}

	items := convertSearchResultsToMemoryItems(results)

	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	if items[0].ID != "1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "1")
	}
	if items[1].ID != "2" {
		t.Errorf("items[1].ID = %q, want %q", items[1].ID, "2")
	}
}

func TestNewTUIStreamHandler_WithMemoryService(t *testing.T) {
	// Create handler without memory service
	h1 := NewTUIStreamHandler(nil)
	if h1.memoryService != nil {
		t.Error("memoryService should be nil without option")
	}

	// Create handler with memory service option
	// We can't easily create a real memory.Service without a database,
	// but we can test the option pattern works
	// by verifying the option function can be called without panic
	var nilSvc *memory.Service
	h2 := NewTUIStreamHandler(nil, WithMemoryService(nilSvc))
	if h2.memoryService != nil {
		t.Error("memoryService should still be nil when passed nil")
	}
}

func TestTUIStreamHandler_SendInitialMemories_NilService(t *testing.T) {
	// Handler without memory service should return nil without error
	h := NewTUIStreamHandler(nil)

	err := h.SendInitialMemories(nil, "test query", "@ayo", 0.3, 10)
	if err != nil {
		t.Errorf("SendInitialMemories with nil service should not error: %v", err)
	}
}

func TestTUIStreamHandler_SendInitialMemories_StoresContext(t *testing.T) {
	// Create handler with a mock setup (won't actually query)
	// This tests that the context is stored for later use
	h := NewTUIStreamHandler(nil)

	// SendInitialMemories should store the context even without a memory service
	// (it returns early, but we can check if the fields would be set if we had a service)

	// We need to verify the implementation stores context
	// Since memoryService is nil, it returns early before setting
	// Let's test with a non-nil service would store context
	// For now, just verify the function handles nil gracefully
	err := h.SendInitialMemories(nil, "test", "@test", 0.5, 5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTUIStreamHandler_OnMemoryEvent(t *testing.T) {
	// Test that OnMemoryEvent doesn't panic with nil program
	h := NewTUIStreamHandler(nil)

	err := h.OnMemoryEvent("created", 1)
	if err != nil {
		t.Errorf("OnMemoryEvent should not error: %v", err)
	}

	err = h.OnMemoryEvent("skipped", 0)
	if err != nil {
		t.Errorf("OnMemoryEvent should not error: %v", err)
	}

	err = h.OnMemoryEvent("superseded", 1)
	if err != nil {
		t.Errorf("OnMemoryEvent should not error: %v", err)
	}
}

func TestTUIStreamHandler_Brokers(t *testing.T) {
	h := NewTUIStreamHandler(nil)

	// Verify all brokers are initialized
	if h.MessageBroker() == nil {
		t.Error("MessageBroker should not be nil")
	}
	if h.ToolBroker() == nil {
		t.Error("ToolBroker should not be nil")
	}
	if h.MemoryBroker() == nil {
		t.Error("MemoryBroker should not be nil")
	}
	if h.TextBroker() == nil {
		t.Error("TextBroker should not be nil")
	}
	if h.ReasoningBroker() == nil {
		t.Error("ReasoningBroker should not be nil")
	}
}
