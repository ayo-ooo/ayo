package messages

import (
	"testing"
)

func TestUserMessage_CacheInvalidation(t *testing.T) {
	msg := NewUserMessage("id-1")
	msg.SetContent("Hello")

	// First render
	r1 := msg.Render(80)
	if r1 == "" {
		t.Fatal("expected non-empty render")
	}
	if msg.IsDirty() {
		t.Error("expected clean after render")
	}

	// Second render should use cache
	r2 := msg.Render(80)
	if r1 != r2 {
		t.Error("expected cached result")
	}

	// Append invalidates
	msg.AppendContent(" world")
	if !msg.IsDirty() {
		t.Error("expected dirty after append")
	}

	// Third render should have new content
	r3 := msg.Render(80)
	if r3 == r1 {
		t.Error("expected new render after append")
	}
}

func TestAssistantMessage_CacheInvalidation(t *testing.T) {
	msg := NewAssistantMessage("id-1", "@ayo")
	msg.SetContent("Hello")

	// First render
	r1 := msg.Render(80)
	if r1 == "" {
		t.Fatal("expected non-empty render")
	}
	if msg.IsDirty() {
		t.Error("expected clean after render")
	}

	// Second render should use cache
	r2 := msg.Render(80)
	if r1 != r2 {
		t.Error("expected cached result")
	}

	// Width change triggers re-render
	r3 := msg.Render(100)
	// Content should be same but rendered differently
	if r3 == "" {
		t.Error("expected non-empty render at new width")
	}

	// Append invalidates
	msg.AppendContent(" world")
	if !msg.IsDirty() {
		t.Error("expected dirty after append")
	}

	// Fourth render should have new content
	r4 := msg.Render(80)
	if r4 == r1 {
		t.Error("expected new render after append")
	}
}

func TestAssistantMessage_StreamingContent(t *testing.T) {
	msg := NewAssistantMessage("id-1", "@ayo")

	// Start with empty
	r1 := msg.Render(80)
	if r1 == "" {
		t.Error("expected label even with empty content")
	}

	// Stream in content
	msg.AppendContent("Hello")
	r2 := msg.Render(80)
	if r2 == r1 {
		t.Error("expected different render after content")
	}

	msg.AppendContent(" world")
	r3 := msg.Render(80)
	if r3 == r2 {
		t.Error("expected different render after more content")
	}

	// Mark complete
	msg.SetComplete(true)
	if !msg.IsComplete() {
		t.Error("expected complete")
	}
}

func TestUserMessage_Content(t *testing.T) {
	msg := NewUserMessage("id-1")
	msg.SetContent("test content")

	if msg.Content() != "test content" {
		t.Errorf("expected 'test content', got %q", msg.Content())
	}

	if msg.Role() != "user" {
		t.Errorf("expected 'user', got %q", msg.Role())
	}

	if msg.ID() != "id-1" {
		t.Errorf("expected 'id-1', got %q", msg.ID())
	}
}

func TestAssistantMessage_AgentHandle(t *testing.T) {
	msg := NewAssistantMessage("id-1", "@ayo")

	if msg.AgentHandle() != "@ayo" {
		t.Errorf("expected '@ayo', got %q", msg.AgentHandle())
	}

	if msg.Role() != "assistant" {
		t.Errorf("expected 'assistant', got %q", msg.Role())
	}
}
