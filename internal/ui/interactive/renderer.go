package interactive

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/alexcabrera/ayo/internal/ui/shared"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// EventRenderer renders events to a terminal.
type EventRenderer interface {
	// Render handles a single event
	Render(event Event) error

	// Flush ensures all buffered output is written
	Flush() error

	// Reset clears any state for new conversation turn
	Reset()
}

// SimpleRenderer implements EventRenderer with basic terminal output.
// It's designed to be simple and non-blocking.
type SimpleRenderer struct {
	out          io.Writer
	glamour      *glamour.TermRenderer
	stream       *StreamRenderer
	mu           sync.Mutex
	width        int
	activeTools  map[string]string // id -> name
	toolStyle    lipgloss.Style
	errorStyle   lipgloss.Style
	successStyle lipgloss.Style
	dimStyle     lipgloss.Style
}

// NewSimpleRenderer creates a new simple event renderer.
func NewSimpleRenderer(out io.Writer, width int) *SimpleRenderer {
	return &SimpleRenderer{
		out:          out,
		glamour:      shared.GetStyledMarkdownRenderer(width),
		stream:       NewStreamRenderer(out, width),
		width:        shared.Clamp(width, 40, 200),
		activeTools:  make(map[string]string),
		toolStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#67e8f9")),
		errorStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171")),
		successStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")),
		dimStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")),
	}
}

// Render handles a single event.
func (r *SimpleRenderer) Render(event Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch event.Type {
	case EventText:
		return r.renderText(event)
	case EventToolStart:
		return r.renderToolStart(event)
	case EventToolProgress:
		return r.renderToolProgress(event)
	case EventToolComplete:
		return r.renderToolComplete(event)
	case EventToolError:
		return r.renderToolError(event)
	case EventThinking:
		return r.renderThinking(event)
	case EventDelegate:
		return r.renderDelegate(event)
	case EventPlanUpdate:
		return r.renderPlanUpdate(event)
	case EventMemory:
		return r.renderMemory(event)
	case EventError:
		return r.renderError(event)
	case EventComplete:
		return r.renderComplete(event)
	}

	return nil
}

func (r *SimpleRenderer) renderText(event Event) error {
	data, ok := event.Data.(*TextData)
	if !ok {
		return nil
	}
	return r.stream.WriteToken(data.Text)
}

func (r *SimpleRenderer) renderToolStart(event Event) error {
	data, ok := event.Data.(*ToolStartData)
	if !ok {
		return nil
	}

	r.activeTools[data.ID] = data.Name

	// Flush any pending text before tool output
	if err := r.stream.Flush(); err != nil {
		return err
	}

	toolName := r.toolStyle.Render(data.Name)
	_, err := fmt.Fprintf(r.out, "\n  ▸ %s: %s\n", toolName, data.Summary)
	return err
}

func (r *SimpleRenderer) renderToolProgress(event Event) error {
	data, ok := event.Data.(*ToolProgressData)
	if !ok {
		return nil
	}

	msg := r.dimStyle.Render(data.Message)
	_, err := fmt.Fprintf(r.out, "    ⋯ %s\n", msg)
	return err
}

func (r *SimpleRenderer) renderToolComplete(event Event) error {
	data, ok := event.Data.(*ToolCompleteData)
	if !ok {
		return nil
	}

	delete(r.activeTools, data.ID)

	icon := "✓"
	style := r.successStyle
	if !data.Success {
		icon = "✗"
		style = r.errorStyle
	}

	summary := style.Render(data.Summary)
	_, err := fmt.Fprintf(r.out, "    └─ %s %s\n", icon, summary)
	return err
}

func (r *SimpleRenderer) renderToolError(event Event) error {
	data, ok := event.Data.(*ToolErrorData)
	if !ok {
		return nil
	}

	delete(r.activeTools, data.ID)

	errMsg := r.errorStyle.Render(data.Error)
	_, err := fmt.Fprintf(r.out, "    └─ ✗ %s\n", errMsg)
	return err
}

func (r *SimpleRenderer) renderThinking(event Event) error {
	data, ok := event.Data.(*ThinkingData)
	if !ok {
		return nil
	}

	if data.Active {
		msg := r.dimStyle.Render(data.Message)
		_, err := fmt.Fprintf(r.out, "  ⋯ %s\n", msg)
		return err
	}
	return nil
}

func (r *SimpleRenderer) renderDelegate(event Event) error {
	data, ok := event.Data.(*DelegateData)
	if !ok {
		return nil
	}

	// Flush any pending text
	if err := r.stream.Flush(); err != nil {
		return err
	}

	if data.Started {
		agent := r.toolStyle.Render("@" + data.Agent)
		_, err := fmt.Fprintf(r.out, "\n  → Delegating to %s: %s\n\n", agent, data.Task)
		return err
	}

	agent := r.toolStyle.Render("@" + data.Agent)
	_, err := fmt.Fprintf(r.out, "\n  ← %s completed\n", agent)
	return err
}

func (r *SimpleRenderer) renderPlanUpdate(event Event) error {
	data, ok := event.Data.(*PlanUpdateData)
	if !ok {
		return nil
	}

	// Flush any pending text
	if err := r.stream.Flush(); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("\n")
	for _, item := range data.Items {
		icon := r.getPlanIcon(item.Status)
		content := item.Content
		if item.Status == PlanItemInProgress {
			content = content + " ← in progress"
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", icon, content))
	}

	_, err := r.out.Write([]byte(sb.String()))
	return err
}

func (r *SimpleRenderer) getPlanIcon(status PlanItemStatus) string {
	switch status {
	case PlanItemCompleted:
		return r.successStyle.Render("⊠")
	case PlanItemInProgress:
		return r.toolStyle.Render("◉")
	default:
		return r.dimStyle.Render("○")
	}
}

func (r *SimpleRenderer) renderMemory(event Event) error {
	data, ok := event.Data.(*MemoryData)
	if !ok {
		return nil
	}

	icon := "📝"
	switch data.Operation {
	case "read":
		icon = "📖"
	case "delete":
		icon = "🗑"
	}

	msg := r.dimStyle.Render(data.Message)
	_, err := fmt.Fprintf(r.out, "  %s %s\n", icon, msg)
	return err
}

func (r *SimpleRenderer) renderError(event Event) error {
	data, ok := event.Data.(*ErrorData)
	if !ok {
		return nil
	}

	// Flush any pending text
	if err := r.stream.Flush(); err != nil {
		return err
	}

	errMsg := r.errorStyle.Render("Error: " + data.Message)
	_, err := fmt.Fprintf(r.out, "\n%s\n", errMsg)
	if data.Details != "" {
		details := r.dimStyle.Render(data.Details)
		fmt.Fprintf(r.out, "%s\n", details)
	}
	return err
}

func (r *SimpleRenderer) renderComplete(event Event) error {
	// Flush any pending text
	return r.stream.Flush()
}

// Flush ensures all buffered output is written.
func (r *SimpleRenderer) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stream.Flush()
}

// Reset clears any state for new conversation turn.
func (r *SimpleRenderer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stream.Reset()
	r.activeTools = make(map[string]string)
}
