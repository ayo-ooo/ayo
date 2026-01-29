package messages

import (
	"strings"
	"testing"
)

func TestReasoningCmp_Lifecycle(t *testing.T) {
	r := NewReasoningCmp()

	// Initially empty
	if !r.IsEmpty() {
		t.Error("expected empty initially")
	}

	// Append content
	r.AppendContent("Thinking about")
	if r.IsEmpty() {
		t.Error("expected not empty after append")
	}
	if r.Content() != "Thinking about" {
		t.Errorf("content = %q, want %q", r.Content(), "Thinking about")
	}

	// More content
	r.AppendContent(" the problem")
	if r.Content() != "Thinking about the problem" {
		t.Errorf("content = %q, want %q", r.Content(), "Thinking about the problem")
	}

	// Complete
	r.SetComplete(true)
	if !r.IsComplete() {
		t.Error("expected complete")
	}

	// Reset
	r.Reset()
	if !r.IsEmpty() {
		t.Error("expected empty after reset")
	}
	if r.IsComplete() {
		t.Error("expected not complete after reset")
	}
}

func TestReasoningCmp_Truncation(t *testing.T) {
	r := NewReasoningCmp()

	// Add more than 5 lines
	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5", "Line 6", "Line 7"}
	r.AppendContent(strings.Join(lines, "\n"))

	rendered := r.Render(80)

	// Should NOT contain Line 1 and Line 2 (first two lines)
	if strings.Contains(rendered, "Line 1") {
		t.Error("should not contain Line 1 after truncation")
	}
	if strings.Contains(rendered, "Line 2") {
		t.Error("should not contain Line 2 after truncation")
	}

	// Should contain the last 5 lines
	if !strings.Contains(rendered, "Line 7") {
		t.Error("should contain Line 7")
	}
	if !strings.Contains(rendered, "Line 3") {
		t.Error("should contain Line 3")
	}
}

func TestReasoningCmp_CacheInvalidation(t *testing.T) {
	r := NewReasoningCmp()
	r.AppendContent("Test content")

	// First render
	r1 := r.Render(80)
	if r1 == "" {
		t.Fatal("expected non-empty render")
	}
	if r.IsDirty() {
		t.Error("expected clean after render")
	}

	// Second render should use cache
	r2 := r.Render(80)
	if r1 != r2 {
		t.Error("expected cached result")
	}

	// Append invalidates
	r.AppendContent(" more")
	if !r.IsDirty() {
		t.Error("expected dirty after append")
	}

	// Third render should have new content
	r3 := r.Render(80)
	if r3 == r1 {
		t.Error("expected different render after append")
	}
}

func TestReasoningCmp_EmptyRender(t *testing.T) {
	r := NewReasoningCmp()

	rendered := r.Render(80)
	if rendered != "" {
		t.Error("expected empty render for empty content")
	}
}

func TestReasoningCmp_Label(t *testing.T) {
	r := NewReasoningCmp()
	r.AppendContent("Test")

	rendered := r.Render(80)
	if !strings.Contains(rendered, "Thinking") {
		t.Error("expected 'Thinking' label in render")
	}
}
