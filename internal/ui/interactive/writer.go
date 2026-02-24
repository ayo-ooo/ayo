package interactive

import (
	"os"
	"time"

	"github.com/alexcabrera/ayo/internal/run"
	"golang.org/x/term"
)

// Writer implements run.StreamWriter using the SimpleRenderer.
// It bridges the streaming output interface with the new interactive renderer.
type Writer struct {
	renderer *SimpleRenderer
}

// NewWriter creates a new interactive writer that renders to stdout.
func NewWriter() *Writer {
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if width <= 0 {
		width = 80
	}
	return &Writer{
		renderer: NewSimpleRenderer(os.Stdout, width),
	}
}

// NewWriterWithRenderer creates a writer with a custom renderer.
func NewWriterWithRenderer(renderer *SimpleRenderer) *Writer {
	return &Writer{renderer: renderer}
}

// WriteText handles streaming text chunks.
func (w *Writer) WriteText(delta string) {
	w.renderer.Render(Event{
		Type: EventText,
		Data: &TextData{Text: delta},
	})
}

// WriteTextDone signals text streaming is complete.
func (w *Writer) WriteTextDone(content string) {
	w.renderer.Render(Event{
		Type: EventComplete,
		Data: &CompleteData{},
	})
}

// WriteReasoning handles thinking/reasoning output.
func (w *Writer) WriteReasoning(delta string) {
	w.renderer.Render(Event{
		Type: EventThinking,
		Data: &ThinkingData{Active: true, Message: delta},
	})
}

// WriteReasoningDone signals reasoning is complete.
func (w *Writer) WriteReasoningDone(content string, duration time.Duration) {
	w.renderer.Render(Event{
		Type: EventThinking,
		Data: &ThinkingData{Active: false},
	})
}

// WriteToolStart handles a tool invocation beginning.
func (w *Writer) WriteToolStart(call run.ToolCall) {
	summary := call.Description
	if summary == "" {
		summary = call.Input
	}
	if len(summary) > 60 {
		summary = summary[:57] + "..."
	}

	w.renderer.Render(Event{
		Type: EventToolStart,
		Data: &ToolStartData{
			ID:      call.ID,
			Name:    call.Name,
			Summary: summary,
		},
	})
}

// WriteToolResult handles a completed tool result.
func (w *Writer) WriteToolResult(result run.ToolResult) {
	if result.Error != "" {
		w.renderer.Render(Event{
			Type: EventToolError,
			Data: &ToolErrorData{
				ID:    result.ID,
				Error: result.Error,
			},
		})
		return
	}

	w.renderer.Render(Event{
		Type: EventToolComplete,
		Data: &ToolCompleteData{
			ID:      result.ID,
			Success: true,
			Summary: summarizeResult(result.Name, result.Output),
		},
	})
}

// WriteAgentStart handles a sub-agent delegation beginning.
func (w *Writer) WriteAgentStart(handle, prompt string) {
	w.renderer.Render(Event{
		Type: EventDelegate,
		Data: &DelegateData{
			Agent:   handle,
			Task:    prompt,
			Started: true,
		},
	})
}

// WriteAgentEnd handles a sub-agent delegation completing.
func (w *Writer) WriteAgentEnd(handle string, duration time.Duration, err error) {
	w.renderer.Render(Event{
		Type: EventDelegate,
		Data: &DelegateData{
			Agent:   handle,
			Started: false,
		},
	})
}

// WriteMemoryEvent handles memory operations.
func (w *Writer) WriteMemoryEvent(event string, count int) {
	w.renderer.Render(Event{
		Type: EventMemory,
		Data: &MemoryData{
			Operation: event,
			Message:   event,
		},
	})
}

// WriteError handles error output.
func (w *Writer) WriteError(err error) {
	w.renderer.Render(Event{
		Type: EventError,
		Data: &ErrorData{Message: err.Error()},
	})
}

// WriteDone signals the response is complete.
func (w *Writer) WriteDone(response string) {
	w.renderer.Flush()
}

// Verify Writer implements run.StreamWriter.
var _ run.StreamWriter = (*Writer)(nil)
