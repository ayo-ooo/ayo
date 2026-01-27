package run

import (
	"encoding/json"
	"fmt"
	"time"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/ui"
)

// PrintStreamHandler wraps the existing UI to implement StreamHandler.
// This maintains backward compatibility with non-TUI mode.
type PrintStreamHandler struct {
	ui             *ui.UI
	spinner        *ui.Spinner
	spinnerActive  bool
	textStarted    bool
	agentHandle    string
	reasoningStart time.Time
}

// NewPrintStreamHandler creates a handler that prints to the terminal.
func NewPrintStreamHandler(u *ui.UI) *PrintStreamHandler {
	return &PrintStreamHandler{ui: u}
}

// NewPrintStreamHandlerWithSpinner creates a handler with debug and depth options.
func NewPrintStreamHandlerWithSpinner(debug bool, depth int, agentHandle string) *PrintStreamHandler {
	u := ui.NewWithDepth(debug, depth)
	spinner := ui.NewSpinnerWithDepth("thinking...", depth)
	spinner.Start()
	return &PrintStreamHandler{
		ui:            u,
		spinner:       spinner,
		spinnerActive: true,
		agentHandle:   agentHandle,
	}
}

func (h *PrintStreamHandler) OnTextDelta(id, text string) error {
	if h.spinnerActive {
		h.spinner.Stop()
		h.spinnerActive = false
	}
	if !h.textStarted {
		h.textStarted = true
		h.ui.PrintAgentResponseHeader(h.agentHandle)
	}
	h.ui.PrintTextDelta(text)
	return nil
}

func (h *PrintStreamHandler) OnTextEnd(id string) error {
	h.ui.PrintTextEnd()
	return nil
}

func (h *PrintStreamHandler) OnReasoningStart(id string) error {
	if h.spinnerActive {
		h.spinner.Stop()
		h.spinnerActive = false
	}
	h.reasoningStart = time.Now()
	h.ui.PrintReasoningStart()
	return nil
}

func (h *PrintStreamHandler) OnReasoningDelta(id, text string) error {
	h.ui.PrintReasoningDelta(text)
	return nil
}

func (h *PrintStreamHandler) OnReasoningEnd(id string, duration time.Duration) error {
	h.ui.PrintReasoningEnd()
	if duration > 0 {
		h.ui.PrintThinkingDone(formatDuration(duration))
	}
	return nil
}

func (h *PrintStreamHandler) OnToolCall(tc fantasy.ToolCallContent) error {
	if h.spinnerActive {
		h.spinner.Stop()
		h.spinnerActive = false
	}
	info := ui.ToolCallInfo{
		Name:  tc.ToolName,
		Input: tc.Input,
	}

	// Extract description and command from input for bash
	if tc.ToolName == "bash" {
		var params struct {
			Command     string `json:"command"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal([]byte(tc.Input), &params); err == nil {
			info.Description = params.Description
			info.Command = params.Command
		}
	}

	h.ui.PrintToolCallStart(info)
	return nil
}

func (h *PrintStreamHandler) OnToolResult(result fantasy.ToolResultContent, duration time.Duration) error {
	info := ui.ToolCallInfo{
		Duration: formatDuration(duration),
		Output:   formatToolResultContent(result),
		Metadata: result.ClientMetadata,
	}

	if result.Result.GetType() == fantasy.ToolResultContentTypeError {
		info.Error = info.Output
	}

	h.ui.PrintToolCallResult(info)
	return nil
}

func (h *PrintStreamHandler) OnAgentStart(handle, prompt string) error {
	h.ui.PrintSubAgentStart(handle, prompt)
	return nil
}

func (h *PrintStreamHandler) OnAgentEnd(handle string, duration time.Duration, err error) error {
	h.ui.PrintSubAgentEnd(handle, formatDuration(duration), err != nil)
	return nil
}

func (h *PrintStreamHandler) OnMemoryEvent(event string, count int) error {
	h.ui.PrintMemoryEvent(ui.MemoryEventType(event))
	return nil
}

func (h *PrintStreamHandler) OnError(err error) error {
	if h.spinnerActive {
		h.spinner.StopWithError("Failed")
		h.spinnerActive = false
	}
	h.ui.PrintError(err.Error())
	return nil
}

// formatDuration formats a duration as a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
