package run

import (
	"encoding/json"
	"time"

	"github.com/alexcabrera/ayo/internal/ui"
)

// PrintWriter implements StreamWriter for non-interactive mode.
// It writes directly to stdout/stderr with spinner management.
type PrintWriter struct {
	ui            *ui.UI
	spinner       *ui.Spinner
	spinnerActive bool
	textStarted   bool
	agentHandle   string
}

// NewPrintWriter creates a writer for non-interactive output.
func NewPrintWriter(agentHandle string, debug bool, depth int) *PrintWriter {
	u := ui.NewWithDepth(debug, depth)
	spinner := ui.NewSpinnerWithDepth("thinking...", depth)
	spinner.Start()
	return &PrintWriter{
		ui:            u,
		spinner:       spinner,
		spinnerActive: true,
		agentHandle:   agentHandle,
	}
}

// NewPrintWriterWithUI creates a writer with an existing UI instance.
func NewPrintWriterWithUI(u *ui.UI, agentHandle string) *PrintWriter {
	spinner := ui.NewSpinnerWithDepth("thinking...", u.Depth())
	spinner.Start()
	return &PrintWriter{
		ui:            u,
		spinner:       spinner,
		spinnerActive: true,
		agentHandle:   agentHandle,
	}
}

func (w *PrintWriter) WriteText(delta string) {
	if w.spinnerActive {
		w.spinner.Stop()
		w.spinnerActive = false
	}
	if !w.textStarted {
		w.textStarted = true
		w.ui.PrintAgentResponseHeader(w.agentHandle)
	}
	w.ui.PrintTextDelta(delta)
}

func (w *PrintWriter) WriteTextDone(content string) {
	w.ui.PrintTextEnd()
}

func (w *PrintWriter) WriteReasoning(delta string) {
	if w.spinnerActive {
		w.spinner.Stop()
		w.spinnerActive = false
	}
	w.ui.PrintReasoningDelta(delta)
}

func (w *PrintWriter) WriteReasoningDone(content string, duration time.Duration) {
	w.ui.PrintReasoningEnd()
	if duration > 0 {
		w.ui.PrintThinkingDone(formatDuration(duration))
	}
}

func (w *PrintWriter) WriteToolStart(call ToolCall) {
	if w.spinnerActive {
		w.spinner.Stop()
		w.spinnerActive = false
	}
	info := ui.ToolCallInfo{
		Name:        call.Name,
		Description: call.Description,
		Command:     call.Command,
		Input:       call.Input,
	}
	w.ui.PrintToolCallStart(info)
}

func (w *PrintWriter) WriteToolResult(result ToolResult) {
	info := ui.ToolCallInfo{
		Name:     result.Name,
		Output:   result.Output,
		Error:    result.Error,
		Duration: formatDuration(result.Duration),
		Metadata: result.Metadata,
	}
	w.ui.PrintToolCallResult(info)
}

func (w *PrintWriter) WriteAgentStart(handle, prompt string) {
	w.ui.PrintSubAgentStart(handle, prompt)
}

func (w *PrintWriter) WriteAgentEnd(handle string, duration time.Duration, err error) {
	w.ui.PrintSubAgentEnd(handle, formatDuration(duration), err != nil)
}

func (w *PrintWriter) WriteMemoryEvent(event string, count int) {
	w.ui.PrintMemoryEvent(ui.MemoryEventType(event))
}

func (w *PrintWriter) WriteError(err error) {
	if w.spinnerActive {
		w.spinner.StopWithError("Failed")
		w.spinnerActive = false
	}
	w.ui.PrintError(err.Error())
}

func (w *PrintWriter) WriteDone(response string) {
	// Nothing special to do for print mode - output is already written
}

// Verify PrintWriter implements StreamWriter
var _ StreamWriter = (*PrintWriter)(nil)

// extractBashParams extracts command and description from bash tool input JSON.
func extractBashParams(input string) (command, description string) {
	var params struct {
		Command     string `json:"command"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(input), &params); err == nil {
		return params.Command, params.Description
	}
	return "", ""
}
