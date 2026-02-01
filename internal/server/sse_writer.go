package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/run"
)

// SSEWriter implements run.StreamWriter by sending Server-Sent Events over HTTP.
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
	mu      sync.Mutex
	closed  bool
	eventID int
}

// NewSSEWriter creates a new SSE writer for the given response.
// Returns nil if the response writer doesn't support flushing.
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	return &SSEWriter{
		w:       w,
		flusher: flusher,
	}
}

// writeEvent sends an SSE event with the given type and data.
func (s *SSEWriter) writeEvent(eventType string, data any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.eventID++

	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fall back to error message
		jsonData = []byte(fmt.Sprintf(`{"error":"marshal failed: %s"}`, err.Error()))
	}

	fmt.Fprintf(s.w, "id: %d\n", s.eventID)
	fmt.Fprintf(s.w, "event: %s\n", eventType)
	fmt.Fprintf(s.w, "data: %s\n\n", jsonData)
	s.flusher.Flush()
}

// SendHeartbeat sends a comment line to keep the connection alive.
func (s *SSEWriter) SendHeartbeat() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	fmt.Fprint(s.w, ": heartbeat\n\n")
	s.flusher.Flush()
}

// Close marks the writer as closed.
func (s *SSEWriter) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
}

// WriteText sends a text delta event.
func (s *SSEWriter) WriteText(delta string) {
	s.writeEvent("text_delta", map[string]string{
		"delta": delta,
	})
}

// WriteTextDone sends a text completion event.
func (s *SSEWriter) WriteTextDone(content string) {
	s.writeEvent("text_done", map[string]string{
		"content": content,
	})
}

// WriteReasoning sends a reasoning delta event.
func (s *SSEWriter) WriteReasoning(delta string) {
	s.writeEvent("reasoning_delta", map[string]string{
		"delta": delta,
	})
}

// WriteReasoningDone sends a reasoning completion event.
func (s *SSEWriter) WriteReasoningDone(content string, duration time.Duration) {
	s.writeEvent("reasoning_done", map[string]any{
		"content":     content,
		"duration_ms": duration.Milliseconds(),
	})
}

// WriteToolStart sends a tool start event.
func (s *SSEWriter) WriteToolStart(call run.ToolCall) {
	s.writeEvent("tool_start", map[string]any{
		"id":          call.ID,
		"name":        call.Name,
		"description": call.Description,
		"command":     call.Command,
		"input":       call.Input,
		"parent_id":   call.ParentID,
	})
}

// WriteToolResult sends a tool result event.
func (s *SSEWriter) WriteToolResult(result run.ToolResult) {
	s.writeEvent("tool_result", map[string]any{
		"id":          result.ID,
		"name":        result.Name,
		"output":      result.Output,
		"error":       result.Error,
		"duration_ms": result.Duration.Milliseconds(),
		"metadata":    result.Metadata,
	})
}

// WriteAgentStart sends an agent start event (for sub-agent calls).
func (s *SSEWriter) WriteAgentStart(handle, prompt string) {
	s.writeEvent("agent_start", map[string]string{
		"handle": handle,
		"prompt": prompt,
	})
}

// WriteAgentEnd sends an agent end event.
func (s *SSEWriter) WriteAgentEnd(handle string, duration time.Duration, err error) {
	data := map[string]any{
		"handle":      handle,
		"duration_ms": duration.Milliseconds(),
	}
	if err != nil {
		data["error"] = err.Error()
	}
	s.writeEvent("agent_end", data)
}

// WriteMemoryEvent sends a memory event.
func (s *SSEWriter) WriteMemoryEvent(event string, count int) {
	s.writeEvent("memory", map[string]any{
		"event": event,
		"count": count,
	})
}

// WriteError sends an error event.
func (s *SSEWriter) WriteError(err error) {
	s.writeEvent("error", map[string]string{
		"message": err.Error(),
	})
}

// WriteDone sends a completion event with the final response.
func (s *SSEWriter) WriteDone(response string) {
	s.writeEvent("done", map[string]string{
		"response": response,
	})
}

// Verify SSEWriter implements run.StreamWriter
var _ run.StreamWriter = (*SSEWriter)(nil)
