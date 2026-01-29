package run

import (
	"fmt"
	"os"
	"time"
)

// EventType identifies the type of stream event.
type EventType int

const (
	EventTextDelta EventType = iota
	EventTextDone
	EventReasoningDelta
	EventReasoningDone
	EventToolStart
	EventToolResult
	EventAgentStart
	EventAgentEnd
	EventMemory
	EventError
	EventDone
)

// StreamEvent is a unified event type for all streaming events.
// This is sent through a channel and forwarded to the TUI.
type StreamEvent struct {
	Type EventType

	// Text/Reasoning content
	Delta    string
	Content  string
	Duration time.Duration

	// Tool events
	Call   *ToolCall
	Result *ToolResult

	// Agent events
	Handle string
	Prompt string
	Err    error

	// Memory events
	MemoryEvent string
	MemoryCount int

	// Final response
	Response string
}

// debug logging
func debugLog(format string, args ...interface{}) {
	f, err := os.OpenFile("/tmp/ayo_stream.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%d: ", time.Now().UnixMilli())
	fmt.Fprintf(f, format, args...)
	fmt.Fprintln(f)
}

// ChannelWriter implements StreamWriter by sending events to a channel.
// This is used in TUI mode - events are read by an EventAggregator
// and forwarded to the Bubble Tea program.
type ChannelWriter struct {
	events chan<- StreamEvent
}

// NewChannelWriter creates a writer that sends events to the given channel.
func NewChannelWriter(events chan<- StreamEvent) *ChannelWriter {
	debugLog("NewChannelWriter created")
	return &ChannelWriter{events: events}
}

func (w *ChannelWriter) WriteText(delta string) {
	debugLog("WriteText: %q", delta)
	w.events <- StreamEvent{Type: EventTextDelta, Delta: delta}
}

func (w *ChannelWriter) WriteTextDone(content string) {
	debugLog("WriteTextDone")
	w.events <- StreamEvent{Type: EventTextDone, Content: content}
}

func (w *ChannelWriter) WriteReasoning(delta string) {
	w.events <- StreamEvent{Type: EventReasoningDelta, Delta: delta}
}

func (w *ChannelWriter) WriteReasoningDone(content string, duration time.Duration) {
	w.events <- StreamEvent{Type: EventReasoningDone, Content: content, Duration: duration}
}

func (w *ChannelWriter) WriteToolStart(call ToolCall) {
	w.events <- StreamEvent{Type: EventToolStart, Call: &call}
}

func (w *ChannelWriter) WriteToolResult(result ToolResult) {
	w.events <- StreamEvent{Type: EventToolResult, Result: &result}
}

func (w *ChannelWriter) WriteAgentStart(handle, prompt string) {
	w.events <- StreamEvent{Type: EventAgentStart, Handle: handle, Prompt: prompt}
}

func (w *ChannelWriter) WriteAgentEnd(handle string, duration time.Duration, err error) {
	w.events <- StreamEvent{Type: EventAgentEnd, Handle: handle, Duration: duration, Err: err}
}

func (w *ChannelWriter) WriteMemoryEvent(event string, count int) {
	w.events <- StreamEvent{Type: EventMemory, MemoryEvent: event, MemoryCount: count}
}

func (w *ChannelWriter) WriteError(err error) {
	w.events <- StreamEvent{Type: EventError, Err: err}
}

func (w *ChannelWriter) WriteDone(response string) {
	w.events <- StreamEvent{Type: EventDone, Response: response}
}

// Verify ChannelWriter implements StreamWriter
var _ StreamWriter = (*ChannelWriter)(nil)
