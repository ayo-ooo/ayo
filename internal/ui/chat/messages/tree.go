// Package messages provides components for rendering chat messages.
package messages

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
)

// ToolCallTree manages a collection of tool calls rendered as a connected tree.
// It handles adding, removing, and toggling expansion of tool call nodes.
type ToolCallTree struct {
	calls []ToolCallCmp
	width int
}

// NewToolCallTree creates a new tool call tree.
func NewToolCallTree() *ToolCallTree {
	return &ToolCallTree{
		calls: make([]ToolCallCmp, 0),
	}
}

// Add appends a tool call to the tree.
func (t *ToolCallTree) Add(call ToolCallCmp) {
	t.calls = append(t.calls, call)
}

// Remove removes a tool call by ID.
func (t *ToolCallTree) Remove(id string) bool {
	for i, call := range t.calls {
		if call.ID() == id {
			t.calls = append(t.calls[:i], t.calls[i+1:]...)
			return true
		}
	}
	return false
}

// Get retrieves a tool call by ID.
func (t *ToolCallTree) Get(id string) ToolCallCmp {
	for _, call := range t.calls {
		if call.ID() == id {
			return call
		}
		// Also check nested calls
		for _, nested := range call.GetNestedToolCalls() {
			if nested.ID() == id {
				return nested
			}
		}
	}
	return nil
}

// ToggleExpand toggles the expanded state of a tool call by ID.
func (t *ToolCallTree) ToggleExpand(id string) bool {
	if call := t.Get(id); call != nil {
		call.ToggleExpanded()
		return true
	}
	return false
}

// SetWidth sets the rendering width for all tool calls.
func (t *ToolCallTree) SetWidth(width int) {
	t.width = width
	for _, call := range t.calls {
		call.SetSize(width, 0)
	}
}

// Count returns the number of top-level tool calls.
func (t *ToolCallTree) Count() int {
	return len(t.calls)
}

// Clear removes all tool calls.
func (t *ToolCallTree) Clear() {
	t.calls = make([]ToolCallCmp, 0)
}

// All returns all top-level tool calls.
func (t *ToolCallTree) All() []ToolCallCmp {
	return t.calls
}

// Render returns the tree-formatted view of all tool calls.
func (t *ToolCallTree) Render() string {
	if len(t.calls) == 0 {
		return ""
	}

	// For a single call, just render it directly
	if len(t.calls) == 1 {
		return t.calls[0].View()
	}

	// For multiple calls, create a tree structure
	var parts []string
	for i, call := range t.calls {
		view := call.View()

		// Add tree branch prefix for all but first
		if i > 0 && len(t.calls) > 1 {
			parts = append(parts, "") // spacing between calls
		}
		parts = append(parts, view)
	}

	return strings.Join(parts, "\n")
}

// RenderAsTree renders all calls as a single connected tree.
// This creates visual tree structure connecting all tool calls.
func (t *ToolCallTree) RenderAsTree() string {
	if len(t.calls) == 0 {
		return ""
	}

	// Create a root node (empty) with tool calls as children
	root := tree.Root("")

	for _, call := range t.calls {
		root.Child(call.View())
	}

	// Apply rounded enumerator
	return root.Enumerator(roundedEnumerator(0, 2)).String()
}

// RenderWithStyle renders the tree with custom styling.
func (t *ToolCallTree) RenderWithStyle(style lipgloss.Style) string {
	content := t.Render()
	if content == "" {
		return ""
	}
	return style.Render(content)
}

// HasPending returns true if any tool call is still pending (no result).
func (t *ToolCallTree) HasPending() bool {
	for _, call := range t.calls {
		if call.Spinning() {
			return true
		}
		// Check nested calls
		for _, nested := range call.GetNestedToolCalls() {
			if nested.Spinning() {
				return true
			}
		}
	}
	return false
}

// CollapseAll collapses all expanded tool calls.
func (t *ToolCallTree) CollapseAll() {
	for _, call := range t.calls {
		call.SetExpanded(false)
	}
}

// ExpandAll expands all tool calls.
func (t *ToolCallTree) ExpandAll() {
	for _, call := range t.calls {
		call.SetExpanded(true)
	}
}

// AutoCollapse collapses all tool calls except those with pending children.
func (t *ToolCallTree) AutoCollapse() {
	for _, call := range t.calls {
		hasPending := false
		for _, nested := range call.GetNestedToolCalls() {
			if nested.Spinning() {
				hasPending = true
				break
			}
		}
		if !hasPending && len(call.GetNestedToolCalls()) > 0 {
			call.SetExpanded(false)
		}
	}
}
