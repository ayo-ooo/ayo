// Package shared provides tool rendering infrastructure used by both TUI and print modes.
package shared

import (
	"encoding/json"
	"strings"
	"time"
)

// ToolState represents the current state of a tool call.
type ToolState int

const (
	// ToolStatePending means the tool is waiting to execute.
	ToolStatePending ToolState = iota
	// ToolStateRunning means the tool is currently executing.
	ToolStateRunning
	// ToolStateSuccess means the tool completed successfully.
	ToolStateSuccess
	// ToolStateError means the tool failed with an error.
	ToolStateError
	// ToolStateCancelled means the tool was cancelled.
	ToolStateCancelled
)

// ToolRenderInput contains all data needed to render a tool call.
type ToolRenderInput struct {
	// Tool identity
	ID       string
	Name     string
	ParentID string // For nested tool calls

	// Raw payloads (JSON strings)
	RawInput    string
	RawMetadata string
	RawOutput   string

	// State
	State    ToolState
	Duration time.Duration

	// Parsed data (set by ParseParams/ParseMetadata)
	Params   any
	Metadata any
}

// ToolRenderOutput contains mode-agnostic rendering data.
// Both TUI and print modes consume this to produce final styled output.
type ToolRenderOutput struct {
	// Label is the display name (e.g., "Bash", "Todo")
	Label string

	// HeaderParams are key display values for the header line
	// e.g., for bash: the command; for todo: "3/5"
	HeaderParams []string

	// State for icon/color selection
	State ToolState

	// Body sections (rendered in order)
	Sections []RenderSection

	// Flags
	IsNested bool
}

// RenderSection represents a distinct section of tool output.
type RenderSection struct {
	// Type indicates how to render this section
	Type SectionType

	// Content is the raw content to render
	Content string

	// MaxLines limits output (0 = default limit)
	MaxLines int
}

// SectionType indicates the content type for rendering decisions.
type SectionType int

const (
	// SectionPlain is plain text output
	SectionPlain SectionType = iota
	// SectionCode is code/command output (monospace, possibly syntax highlighted)
	SectionCode
	// SectionJSON is JSON output (syntax highlighted)
	SectionJSON
	// SectionMarkdown is markdown content
	SectionMarkdown
	// SectionTodos is a todo list
	SectionTodos
	// SectionError is error content
	SectionError
	// SectionSubAgent is a sub-agent conversation (prompt + response)
	SectionSubAgent
)

// ToolRenderer defines how a tool produces render output.
type ToolRenderer interface {
	// Name returns the tool name this renderer handles.
	Name() string

	// Render produces mode-agnostic render output from input.
	Render(input ToolRenderInput) ToolRenderOutput
}

// toolRendererRegistry holds all registered renderers.
var toolRendererRegistry = map[string]ToolRenderer{}

// RegisterToolRenderer registers a renderer for a tool name.
func RegisterToolRenderer(r ToolRenderer) {
	if r == nil {
		return
	}
	toolRendererRegistry[r.Name()] = r
}

// GetToolRenderer returns the renderer for a tool, or the generic fallback.
func GetToolRenderer(name string) ToolRenderer {
	if r, ok := toolRendererRegistry[name]; ok {
		return r
	}
	return &GenericRenderer{}
}

// RenderTool is a convenience function that looks up the renderer and renders.
func RenderTool(input ToolRenderInput) ToolRenderOutput {
	return GetToolRenderer(input.Name).Render(input)
}

// GenericRenderer handles unknown tools with basic display.
type GenericRenderer struct{}

// Name returns empty string (fallback).
func (g *GenericRenderer) Name() string { return "" }

// Render produces generic output for unknown tools.
func (g *GenericRenderer) Render(input ToolRenderInput) ToolRenderOutput {
	out := ToolRenderOutput{
		Label:    PrettifyToolName(input.Name),
		State:    input.State,
		IsNested: input.ParentID != "",
	}

	// Show raw output if available
	if input.RawOutput != "" {
		out.Sections = append(out.Sections, RenderSection{
			Type:     SectionPlain,
			Content:  input.RawOutput,
			MaxLines: 10,
		})
	}

	return out
}

// PrettifyToolName converts a tool name into a display-friendly label.
func PrettifyToolName(name string) string {
	if name == "" {
		return "Tool"
	}
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// ParseJSON is a helper to unmarshal JSON into a target struct.
func ParseJSON(data string, target any) error {
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), target)
}
