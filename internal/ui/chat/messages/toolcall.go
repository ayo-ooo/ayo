// Package messages provides components for rendering chat messages,
// tool calls, and other content in the TUI. It follows the Crush
// pattern of interface-based components with a renderer registry.
package messages

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexcabrera/ayo/internal/ui/chat/layout"
)

// ToolCall represents the data for a tool invocation.
type ToolCall struct {
	ID       string
	Name     string
	Input    string
	Finished bool
}

// ToolResult represents the result of a tool invocation.
type ToolResult struct {
	ToolCallID string
	Name       string
	Content    string
	IsError    bool
	Metadata   string
	MIMEType   string
	Data       string
}

// ToolCallCmp defines the interface for tool call display components.
// It manages the display of tool execution including pending states,
// results, errors, and nested tool calls.
type ToolCallCmp interface {
	layout.Model
	layout.Sizeable
	layout.Focusable

	// ID returns the unique identifier for this tool call component.
	ID() string

	// GetToolCall returns the tool call data.
	GetToolCall() ToolCall

	// SetToolCall updates the tool call data.
	SetToolCall(tc ToolCall)

	// GetToolResult returns the tool result data.
	GetToolResult() ToolResult

	// SetToolResult updates the tool result.
	SetToolResult(result ToolResult)

	// GetNestedToolCalls returns nested tool calls for hierarchical display.
	GetNestedToolCalls() []ToolCallCmp

	// SetNestedToolCalls sets the nested tool calls.
	SetNestedToolCalls(calls []ToolCallCmp)

	// SetIsNested marks this as a nested tool call.
	SetIsNested(isNested bool)

	// Spinning returns whether the loading animation should be active.
	Spinning() bool

	// SetCancelled marks the tool call as cancelled.
	SetCancelled()

	// ParentMessageID returns the ID of the message that owns this tool call.
	ParentMessageID() string

	// SetPermissionRequested marks that permission was requested.
	SetPermissionRequested()

	// SetPermissionGranted marks that permission was granted.
	SetPermissionGranted()

	// IsExpanded returns whether nested content is expanded.
	IsExpanded() bool

	// ToggleExpanded toggles the expanded state.
	ToggleExpanded()

	// SetExpanded sets the expanded state.
	SetExpanded(expanded bool)
}

// ToolCallOption provides functional options for configuring tool call components.
type ToolCallOption func(*toolCallCmp)

// WithToolCallResult sets the initial tool result.
func WithToolCallResult(result ToolResult) ToolCallOption {
	return func(t *toolCallCmp) {
		t.result = result
	}
}

// WithToolCallCancelled marks the tool call as cancelled.
func WithToolCallCancelled() ToolCallOption {
	return func(t *toolCallCmp) {
		t.cancelled = true
	}
}

// WithToolCallNested marks this as a nested tool call.
func WithToolCallNested(isNested bool) ToolCallOption {
	return func(t *toolCallCmp) {
		t.isNested = isNested
	}
}

// WithToolCallNestedCalls sets the nested tool calls.
func WithToolCallNestedCalls(calls []ToolCallCmp) ToolCallOption {
	return func(t *toolCallCmp) {
		t.nestedToolCalls = calls
	}
}

// WithToolPermissionRequested marks permission as requested.
func WithToolPermissionRequested() ToolCallOption {
	return func(t *toolCallCmp) {
		t.permissionRequested = true
	}
}

// WithToolPermissionGranted marks permission as granted.
func WithToolPermissionGranted() ToolCallOption {
	return func(t *toolCallCmp) {
		t.permissionGranted = true
	}
}

// toolCallCmp implements the ToolCallCmp interface.
type toolCallCmp struct {
	width    int
	height   int
	focused  bool
	isNested bool
	expanded bool // B.03: Tracks if nested content is expanded

	parentMessageID     string
	call                ToolCall
	result              ToolResult
	cancelled           bool
	permissionRequested bool
	permissionGranted   bool

	spinning        bool
	nestedToolCalls []ToolCallCmp
}

// NewToolCallCmp creates a new tool call component.
func NewToolCallCmp(parentMessageID string, tc ToolCall, opts ...ToolCallOption) ToolCallCmp {
	t := &toolCallCmp{
		parentMessageID: parentMessageID,
		call:            tc,
		expanded:        true, // Default to expanded
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// ID returns the tool call ID.
func (t *toolCallCmp) ID() string {
	return t.call.ID
}

// Init initializes the component.
func (t *toolCallCmp) Init() tea.Cmd {
	t.spinning = t.shouldSpin()
	return nil
}

// shouldSpin returns whether the spinner should be active.
func (t *toolCallCmp) shouldSpin() bool {
	return t.result.ToolCallID == "" && !t.cancelled
}

// Update handles messages.
func (t *toolCallCmp) Update(msg tea.Msg) (layout.Model, tea.Cmd) {
	return t, nil
}

// View renders the component using the registered renderer.
func (t *toolCallCmp) View() string {
	r := registry.lookup(t.call.Name)
	return r.Render(t)
}

// Focus gives the component focus.
func (t *toolCallCmp) Focus() tea.Cmd {
	t.focused = true
	return nil
}

// Blur removes focus from the component.
func (t *toolCallCmp) Blur() tea.Cmd {
	t.focused = false
	return nil
}

// IsFocused returns whether the component has focus.
func (t *toolCallCmp) IsFocused() bool {
	return t.focused
}

// SetSize sets the component dimensions.
func (t *toolCallCmp) SetSize(width, height int) tea.Cmd {
	t.width = width
	t.height = height
	return nil
}

// GetSize returns the component dimensions.
func (t *toolCallCmp) GetSize() (int, int) {
	return t.width, t.height
}

// GetToolCall returns the tool call data.
func (t *toolCallCmp) GetToolCall() ToolCall {
	return t.call
}

// SetToolCall updates the tool call data.
func (t *toolCallCmp) SetToolCall(tc ToolCall) {
	t.call = tc
}

// GetToolResult returns the tool result.
func (t *toolCallCmp) GetToolResult() ToolResult {
	return t.result
}

// SetToolResult updates the tool result.
func (t *toolCallCmp) SetToolResult(result ToolResult) {
	t.result = result
	t.spinning = t.shouldSpin()
}

// GetNestedToolCalls returns nested tool calls.
func (t *toolCallCmp) GetNestedToolCalls() []ToolCallCmp {
	return t.nestedToolCalls
}

// SetNestedToolCalls sets nested tool calls.
func (t *toolCallCmp) SetNestedToolCalls(calls []ToolCallCmp) {
	t.nestedToolCalls = calls
}

// SetIsNested marks this as nested.
func (t *toolCallCmp) SetIsNested(isNested bool) {
	t.isNested = isNested
}

// Spinning returns the spinner state.
func (t *toolCallCmp) Spinning() bool {
	return t.spinning
}

// SetCancelled marks as cancelled.
func (t *toolCallCmp) SetCancelled() {
	t.cancelled = true
	t.spinning = false
}

// ParentMessageID returns the parent message ID.
func (t *toolCallCmp) ParentMessageID() string {
	return t.parentMessageID
}

// SetPermissionRequested marks permission as requested.
func (t *toolCallCmp) SetPermissionRequested() {
	t.permissionRequested = true
}

// SetPermissionGranted marks permission as granted.
func (t *toolCallCmp) SetPermissionGranted() {
	t.permissionGranted = true
}

// IsExpanded returns whether nested content is expanded.
func (t *toolCallCmp) IsExpanded() bool {
	return t.expanded
}

// ToggleExpanded toggles the expanded state.
func (t *toolCallCmp) ToggleExpanded() {
	t.expanded = !t.expanded
}

// SetExpanded sets the expanded state.
func (t *toolCallCmp) SetExpanded(expanded bool) {
	t.expanded = expanded
}

// textWidth returns the available width for text content.
func (t *toolCallCmp) textWidth() int {
	if t.width <= 0 {
		return 80
	}
	return t.width
}
