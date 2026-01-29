package messages

import (
	"testing"
)

func TestMessageList_AddAndGet(t *testing.T) {
	list := NewMessageList()

	if list.Count() != 0 {
		t.Errorf("expected empty list, got %d", list.Count())
	}

	msg1 := NewUserMessage("id-1")
	msg1.SetContent("Hello")
	list.Add(msg1)

	if list.Count() != 1 {
		t.Errorf("expected 1 message, got %d", list.Count())
	}

	msg2 := NewAssistantMessage("id-2", "@ayo")
	msg2.SetContent("Hi there")
	list.Add(msg2)

	if list.Count() != 2 {
		t.Errorf("expected 2 messages, got %d", list.Count())
	}
}

func TestMessageList_GetByID(t *testing.T) {
	list := NewMessageList()

	msg1 := NewUserMessage("id-1")
	msg1.SetContent("Hello")
	list.Add(msg1)

	msg2 := NewAssistantMessage("id-2", "@ayo")
	msg2.SetContent("Hi there")
	list.Add(msg2)

	// Find existing
	found := list.GetByID("id-1")
	if found == nil {
		t.Fatal("expected to find id-1")
	}
	if found.ID() != "id-1" {
		t.Errorf("expected id-1, got %s", found.ID())
	}

	// Find second
	found2 := list.GetByID("id-2")
	if found2 == nil {
		t.Fatal("expected to find id-2")
	}
	if found2.Role() != "assistant" {
		t.Errorf("expected assistant, got %s", found2.Role())
	}

	// Not found
	notFound := list.GetByID("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent ID")
	}
}

func TestMessageList_GetLast(t *testing.T) {
	list := NewMessageList()

	// Empty list
	if list.GetLast() != nil {
		t.Error("expected nil for empty list")
	}

	msg1 := NewUserMessage("id-1")
	list.Add(msg1)

	last := list.GetLast()
	if last == nil || last.ID() != "id-1" {
		t.Error("expected id-1 as last")
	}

	msg2 := NewAssistantMessage("id-2", "@ayo")
	list.Add(msg2)

	last = list.GetLast()
	if last == nil || last.ID() != "id-2" {
		t.Error("expected id-2 as last after adding second message")
	}
}

func TestMessageList_Render(t *testing.T) {
	list := NewMessageList()

	// Empty list renders empty
	if list.Render(80) != "" {
		t.Error("expected empty render for empty list")
	}

	msg1 := NewUserMessage("id-1")
	msg1.SetContent("Hello")
	list.Add(msg1)

	r1 := list.Render(80)
	if r1 == "" {
		t.Error("expected non-empty render with one message")
	}

	msg2 := NewAssistantMessage("id-2", "@ayo")
	msg2.SetContent("Hi there")
	list.Add(msg2)

	r2 := list.Render(80)
	if len(r2) <= len(r1) {
		t.Error("expected longer render with two messages")
	}
}

func TestMessageList_Clear(t *testing.T) {
	list := NewMessageList()

	msg := NewUserMessage("id-1")
	list.Add(msg)

	if list.Count() != 1 {
		t.Fatal("expected 1 message before clear")
	}

	list.Clear()

	if list.Count() != 0 {
		t.Error("expected 0 messages after clear")
	}
}

func TestMessageList_All(t *testing.T) {
	list := NewMessageList()

	msg1 := NewUserMessage("id-1")
	msg2 := NewUserMessage("id-2")
	list.Add(msg1)
	list.Add(msg2)

	all := list.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(all))
	}

	// Verify it's a copy (modifying shouldn't affect list)
	all[0] = nil
	if list.GetByID("id-1") == nil {
		t.Error("modifying All() result should not affect list")
	}
}
