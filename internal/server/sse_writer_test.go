package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/run"
)

func TestSSEWriterCreation(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	if sse == nil {
		t.Fatal("expected non-nil SSEWriter")
	}

	// Check headers
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Error("expected text/event-stream content type")
	}
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Error("expected no-cache header")
	}
}

func TestSSEWriterTextEvents(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteText("Hello ")
	sse.WriteText("World")
	sse.WriteTextDone("Hello World")

	events := parseSSEEvents(t, w.Body.String())

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	// Check first text delta
	if events[0].Type != "text_delta" {
		t.Errorf("expected text_delta, got %s", events[0].Type)
	}
	if events[0].Data["delta"] != "Hello " {
		t.Errorf("expected 'Hello ', got %q", events[0].Data["delta"])
	}

	// Check text done
	if events[2].Type != "text_done" {
		t.Errorf("expected text_done, got %s", events[2].Type)
	}
	if events[2].Data["content"] != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", events[2].Data["content"])
	}
}

func TestSSEWriterReasoningEvents(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteReasoning("thinking...")
	sse.WriteReasoningDone("full reasoning", 5*time.Second)

	events := parseSSEEvents(t, w.Body.String())

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].Type != "reasoning_delta" {
		t.Errorf("expected reasoning_delta, got %s", events[0].Type)
	}

	if events[1].Type != "reasoning_done" {
		t.Errorf("expected reasoning_done, got %s", events[1].Type)
	}
	if events[1].Data["duration_ms"] != float64(5000) {
		t.Errorf("expected 5000ms duration, got %v", events[1].Data["duration_ms"])
	}
}

func TestSSEWriterToolEvents(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteToolStart(run.ToolCall{
		ID:          "tool-1",
		Name:        "bash",
		Description: "Running tests",
		Command:     "go test ./...",
	})

	sse.WriteToolResult(run.ToolResult{
		ID:       "tool-1",
		Name:     "bash",
		Output:   "PASS",
		Duration: 2 * time.Second,
	})

	events := parseSSEEvents(t, w.Body.String())

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Check tool start
	if events[0].Type != "tool_start" {
		t.Errorf("expected tool_start, got %s", events[0].Type)
	}
	if events[0].Data["name"] != "bash" {
		t.Errorf("expected 'bash', got %q", events[0].Data["name"])
	}
	if events[0].Data["command"] != "go test ./..." {
		t.Errorf("expected 'go test ./...', got %q", events[0].Data["command"])
	}

	// Check tool result
	if events[1].Type != "tool_result" {
		t.Errorf("expected tool_result, got %s", events[1].Type)
	}
	if events[1].Data["output"] != "PASS" {
		t.Errorf("expected 'PASS', got %q", events[1].Data["output"])
	}
}

func TestSSEWriterAgentEvents(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteAgentStart("@crush", "fix the bug")
	sse.WriteAgentEnd("@crush", 10*time.Second, nil)

	events := parseSSEEvents(t, w.Body.String())

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].Type != "agent_start" {
		t.Errorf("expected agent_start, got %s", events[0].Type)
	}
	if events[0].Data["handle"] != "@crush" {
		t.Errorf("expected '@crush', got %q", events[0].Data["handle"])
	}

	if events[1].Type != "agent_end" {
		t.Errorf("expected agent_end, got %s", events[1].Type)
	}
	if _, ok := events[1].Data["error"]; ok {
		t.Error("expected no error field")
	}
}

func TestSSEWriterAgentEndWithError(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteAgentEnd("@crush", 5*time.Second, errors.New("agent failed"))

	events := parseSSEEvents(t, w.Body.String())

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Data["error"] != "agent failed" {
		t.Errorf("expected 'agent failed', got %q", events[0].Data["error"])
	}
}

func TestSSEWriterMemoryEvent(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteMemoryEvent("stored", 3)

	events := parseSSEEvents(t, w.Body.String())

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Type != "memory" {
		t.Errorf("expected memory, got %s", events[0].Type)
	}
	if events[0].Data["event"] != "stored" {
		t.Errorf("expected 'stored', got %q", events[0].Data["event"])
	}
	if events[0].Data["count"] != float64(3) {
		t.Errorf("expected 3, got %v", events[0].Data["count"])
	}
}

func TestSSEWriterErrorEvent(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteError(errors.New("something went wrong"))

	events := parseSSEEvents(t, w.Body.String())

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Type != "error" {
		t.Errorf("expected error, got %s", events[0].Type)
	}
	if events[0].Data["message"] != "something went wrong" {
		t.Errorf("expected error message, got %q", events[0].Data["message"])
	}
}

func TestSSEWriterDoneEvent(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteDone("Full response here")

	events := parseSSEEvents(t, w.Body.String())

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Type != "done" {
		t.Errorf("expected done, got %s", events[0].Type)
	}
	if events[0].Data["response"] != "Full response here" {
		t.Errorf("expected response, got %q", events[0].Data["response"])
	}
}

func TestSSEWriterHeartbeat(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.SendHeartbeat()

	body := w.Body.String()
	if !strings.Contains(body, ": heartbeat") {
		t.Errorf("expected heartbeat comment, got %q", body)
	}
}

func TestSSEWriterClose(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteText("before close")
	sse.Close()
	sse.WriteText("after close") // Should be ignored

	events := parseSSEEvents(t, w.Body.String())

	if len(events) != 1 {
		t.Fatalf("expected 1 event (writes after close should be ignored), got %d", len(events))
	}
}

func TestSSEWriterEventIDs(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	sse.WriteText("one")
	sse.WriteText("two")
	sse.WriteText("three")

	events := parseSSEEvents(t, w.Body.String())

	// Check that event IDs are sequential
	for i, event := range events {
		expectedID := i + 1
		if event.ID != expectedID {
			t.Errorf("expected event ID %d, got %d", expectedID, event.ID)
		}
	}
}

// sseEvent represents a parsed SSE event for testing.
type sseEvent struct {
	ID   int
	Type string
	Data map[string]any
}

// parseSSEEvents parses SSE events from a response body.
func parseSSEEvents(t *testing.T, body string) []sseEvent {
	t.Helper()

	var events []sseEvent
	scanner := bufio.NewScanner(strings.NewReader(body))

	var currentEvent sseEvent
	currentEvent.Data = make(map[string]any)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = end of event
			if currentEvent.Type != "" {
				events = append(events, currentEvent)
			}
			currentEvent = sseEvent{Data: make(map[string]any)}
			continue
		}

		if strings.HasPrefix(line, "id: ") {
			var id int
			if _, err := fmt.Sscanf(line, "id: %d", &id); err == nil {
				currentEvent.ID = id
			}
		} else if strings.HasPrefix(line, "event: ") {
			currentEvent.Type = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if err := json.Unmarshal([]byte(data), &currentEvent.Data); err != nil {
				t.Errorf("failed to parse event data: %v", err)
			}
		}
		// Ignore comment lines (: ...)
	}

	return events
}
