package interactive

import (
	"bytes"
	"testing"
	"time"
)

func TestEventTypes(t *testing.T) {
	// Verify all event types are distinct
	types := []EventType{
		EventText, EventToolStart, EventToolProgress, EventToolComplete,
		EventToolError, EventThinking, EventDelegate, EventPlanUpdate,
		EventMemory, EventError, EventComplete,
	}

	seen := make(map[EventType]bool)
	for _, et := range types {
		if seen[et] {
			t.Errorf("duplicate event type: %s", et)
		}
		seen[et] = true
	}
}

func TestNewEvent(t *testing.T) {
	before := time.Now()
	event := NewEvent(EventText, &TextData{Text: "hello"})
	after := time.Now()

	if event.Type != EventText {
		t.Errorf("expected type EventText, got %s", event.Type)
	}

	if event.Timestamp.Before(before) || event.Timestamp.After(after) {
		t.Error("timestamp not within expected range")
	}

	data, ok := event.Data.(*TextData)
	if !ok {
		t.Fatal("expected TextData")
	}
	if data.Text != "hello" {
		t.Errorf("expected text 'hello', got %q", data.Text)
	}
}

func TestNewTextEvent(t *testing.T) {
	event := NewTextEvent("hello world")

	if event.Type != EventText {
		t.Errorf("expected EventText, got %s", event.Type)
	}

	data := event.Data.(*TextData)
	if data.Text != "hello world" {
		t.Errorf("expected 'hello world', got %q", data.Text)
	}
}

func TestNewToolStartEvent(t *testing.T) {
	event := NewToolStartEvent("id-123", "grep", "searching for 'config'")

	if event.Type != EventToolStart {
		t.Errorf("expected EventToolStart, got %s", event.Type)
	}

	data := event.Data.(*ToolStartData)
	if data.ID != "id-123" {
		t.Errorf("expected ID 'id-123', got %q", data.ID)
	}
	if data.Name != "grep" {
		t.Errorf("expected name 'grep', got %q", data.Name)
	}
	if data.Summary != "searching for 'config'" {
		t.Errorf("expected summary 'searching for config', got %q", data.Summary)
	}
}

func TestNewToolCompleteEvent(t *testing.T) {
	event := NewToolCompleteEvent("id-123", "grep", "found 5 matches", true)

	if event.Type != EventToolComplete {
		t.Errorf("expected EventToolComplete, got %s", event.Type)
	}

	data := event.Data.(*ToolCompleteData)
	if data.ID != "id-123" {
		t.Errorf("expected ID 'id-123', got %q", data.ID)
	}
	if !data.Success {
		t.Error("expected success=true")
	}
}

func TestNewToolErrorEvent(t *testing.T) {
	event := NewToolErrorEvent("id-123", "bash", "command failed")

	if event.Type != EventToolError {
		t.Errorf("expected EventToolError, got %s", event.Type)
	}

	data := event.Data.(*ToolErrorData)
	if data.Error != "command failed" {
		t.Errorf("expected error 'command failed', got %q", data.Error)
	}
}

func TestNewErrorEvent(t *testing.T) {
	event := NewErrorEvent("connection failed", "timeout after 30s")

	if event.Type != EventError {
		t.Errorf("expected EventError, got %s", event.Type)
	}

	data := event.Data.(*ErrorData)
	if data.Message != "connection failed" {
		t.Errorf("expected message 'connection failed', got %q", data.Message)
	}
	if data.Details != "timeout after 30s" {
		t.Errorf("expected details 'timeout after 30s', got %q", data.Details)
	}
}

func TestNewCompleteEvent(t *testing.T) {
	event := NewCompleteEvent(1000, 5*time.Second)

	if event.Type != EventComplete {
		t.Errorf("expected EventComplete, got %s", event.Type)
	}

	data := event.Data.(*CompleteData)
	if data.TokensUsed != 1000 {
		t.Errorf("expected tokens 1000, got %d", data.TokensUsed)
	}
	if data.Duration != 5*time.Second {
		t.Errorf("expected duration 5s, got %s", data.Duration)
	}
}

func TestSimpleRenderer_RenderText(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	event := NewTextEvent("hello\n")
	if err := r.Render(event); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if out.Len() == 0 {
		t.Error("expected output")
	}
}

func TestSimpleRenderer_RenderToolStart(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	event := NewToolStartEvent("id-1", "grep", "searching files")
	if err := r.Render(event); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected output")
	}

	// Should track active tool
	if _, ok := r.activeTools["id-1"]; !ok {
		t.Error("expected tool to be tracked")
	}
}

func TestSimpleRenderer_RenderToolComplete(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	// Start tool first
	r.activeTools["id-1"] = "grep"

	event := NewToolCompleteEvent("id-1", "grep", "5 matches found", true)
	if err := r.Render(event); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected output")
	}

	// Should remove active tool
	if _, ok := r.activeTools["id-1"]; ok {
		t.Error("expected tool to be removed from active")
	}
}

func TestSimpleRenderer_RenderToolError(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	r.activeTools["id-1"] = "bash"

	event := NewToolErrorEvent("id-1", "bash", "command failed")
	if err := r.Render(event); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected output")
	}

	// Should remove active tool
	if _, ok := r.activeTools["id-1"]; ok {
		t.Error("expected tool to be removed from active")
	}
}

func TestSimpleRenderer_RenderError(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	event := NewErrorEvent("API error", "rate limit exceeded")
	if err := r.Render(event); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected output")
	}
}

func TestSimpleRenderer_RenderComplete(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	// Write some text first (will be buffered)
	_ = r.Render(NewTextEvent("partial"))

	// Complete should flush
	event := NewCompleteEvent(100, time.Second)
	if err := r.Render(event); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should have flushed the partial text
	if out.Len() == 0 {
		t.Error("expected output after complete")
	}
}

func TestSimpleRenderer_Reset(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	r.activeTools["id-1"] = "grep"

	r.Reset()

	if len(r.activeTools) != 0 {
		t.Error("expected activeTools to be cleared")
	}
}

func TestSimpleRenderer_Flush(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	// Write partial text
	_ = r.Render(NewTextEvent("partial"))

	if out.Len() != 0 {
		t.Error("expected no output before flush")
	}

	if err := r.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if out.Len() == 0 {
		t.Error("expected output after flush")
	}
}

func TestSimpleRenderer_PlanUpdate(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	event := NewEvent(EventPlanUpdate, &PlanUpdateData{
		Items: []PlanItem{
			{ID: "1", Content: "Analyze code", Status: PlanItemCompleted},
			{ID: "2", Content: "Fix bug", Status: PlanItemInProgress},
			{ID: "3", Content: "Run tests", Status: PlanItemPending},
		},
	})

	if err := r.Render(event); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if out.Len() == 0 {
		t.Error("expected output")
	}
}

func TestSimpleRenderer_Delegate(t *testing.T) {
	var out bytes.Buffer
	r := NewSimpleRenderer(&out, 80)

	// Start delegation
	startEvent := NewEvent(EventDelegate, &DelegateData{
		Agent:   "crush",
		Task:    "implement feature",
		Started: true,
	})

	if err := r.Render(startEvent); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if out.Len() == 0 {
		t.Error("expected output for delegation start")
	}

	out.Reset()

	// Complete delegation
	endEvent := NewEvent(EventDelegate, &DelegateData{
		Agent:   "crush",
		Started: false,
		Result:  "feature implemented",
	})

	if err := r.Render(endEvent); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if out.Len() == 0 {
		t.Error("expected output for delegation complete")
	}
}
