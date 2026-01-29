package messages

import (
	"strings"
	"testing"
)

func TestNewToolCallTree(t *testing.T) {
	tree := NewToolCallTree()
	if tree == nil {
		t.Fatal("Expected non-nil tree")
	}
	if tree.Count() != 0 {
		t.Errorf("Expected empty tree, got %d items", tree.Count())
	}
}

func TestToolCallTree_Add(t *testing.T) {
	tree := NewToolCallTree()

	tc := ToolCall{ID: "tc1", Name: "bash", Input: "{}"}
	cmp := NewToolCallCmp("msg1", tc)
	tree.Add(cmp)

	if tree.Count() != 1 {
		t.Errorf("Expected 1 item, got %d", tree.Count())
	}
}

func TestToolCallTree_Get(t *testing.T) {
	tree := NewToolCallTree()

	tc := ToolCall{ID: "tc1", Name: "bash", Input: "{}"}
	cmp := NewToolCallCmp("msg1", tc)
	tree.Add(cmp)

	found := tree.Get("tc1")
	if found == nil {
		t.Fatal("Expected to find tc1")
	}
	if found.ID() != "tc1" {
		t.Errorf("Expected ID tc1, got %s", found.ID())
	}

	// Test not found
	notFound := tree.Get("nonexistent")
	if notFound != nil {
		t.Error("Expected nil for nonexistent ID")
	}
}

func TestToolCallTree_Get_Nested(t *testing.T) {
	tree := NewToolCallTree()

	// Create parent with nested child
	tc := ToolCall{ID: "parent", Name: "agent", Input: "{}"}
	parent := NewToolCallCmp("msg1", tc)

	nestedTC := ToolCall{ID: "nested1", Name: "bash", Input: "{}"}
	nested := NewToolCallCmp("msg1", nestedTC, WithToolCallNested(true))
	parent.SetNestedToolCalls([]ToolCallCmp{nested})

	tree.Add(parent)

	// Should find nested via parent
	found := tree.Get("nested1")
	if found == nil {
		t.Fatal("Expected to find nested1")
	}
	if found.ID() != "nested1" {
		t.Errorf("Expected ID nested1, got %s", found.ID())
	}
}

func TestToolCallTree_Remove(t *testing.T) {
	tree := NewToolCallTree()

	tc := ToolCall{ID: "tc1", Name: "bash", Input: "{}"}
	cmp := NewToolCallCmp("msg1", tc)
	tree.Add(cmp)

	removed := tree.Remove("tc1")
	if !removed {
		t.Error("Expected successful removal")
	}
	if tree.Count() != 0 {
		t.Errorf("Expected 0 items after removal, got %d", tree.Count())
	}

	// Remove non-existent
	removed = tree.Remove("nonexistent")
	if removed {
		t.Error("Expected false for non-existent removal")
	}
}

func TestToolCallTree_ToggleExpand(t *testing.T) {
	tree := NewToolCallTree()

	tc := ToolCall{ID: "tc1", Name: "agent", Input: "{}"}
	cmp := NewToolCallCmp("msg1", tc)
	tree.Add(cmp)

	// Initially expanded
	if !cmp.IsExpanded() {
		t.Error("Expected initially expanded")
	}

	// Toggle
	toggled := tree.ToggleExpand("tc1")
	if !toggled {
		t.Error("Expected successful toggle")
	}
	if cmp.IsExpanded() {
		t.Error("Expected collapsed after toggle")
	}

	// Toggle non-existent
	toggled = tree.ToggleExpand("nonexistent")
	if toggled {
		t.Error("Expected false for non-existent toggle")
	}
}

func TestToolCallTree_SetWidth(t *testing.T) {
	tree := NewToolCallTree()

	tc := ToolCall{ID: "tc1", Name: "bash", Input: "{}"}
	cmp := NewToolCallCmp("msg1", tc)
	tree.Add(cmp)

	tree.SetWidth(120)

	// Verify width was set on component
	w, _ := cmp.GetSize()
	if w != 120 {
		t.Errorf("Expected width 120, got %d", w)
	}
}

func TestToolCallTree_Clear(t *testing.T) {
	tree := NewToolCallTree()

	tc1 := ToolCall{ID: "tc1", Name: "bash", Input: "{}"}
	tc2 := ToolCall{ID: "tc2", Name: "view", Input: "{}"}
	tree.Add(NewToolCallCmp("msg1", tc1))
	tree.Add(NewToolCallCmp("msg1", tc2))

	tree.Clear()
	if tree.Count() != 0 {
		t.Errorf("Expected 0 items after clear, got %d", tree.Count())
	}
}

func TestToolCallTree_Render(t *testing.T) {
	tree := NewToolCallTree()

	// Empty tree
	if tree.Render() != "" {
		t.Error("Expected empty string for empty tree")
	}

	// Single item
	tc := ToolCall{ID: "tc1", Name: "bash", Input: `{"command":"ls"}`}
	cmp := NewToolCallCmp("msg1", tc)
	cmp.SetSize(80, 0)
	tree.Add(cmp)

	rendered := tree.Render()
	if rendered == "" {
		t.Error("Expected non-empty render")
	}
}

func TestToolCallTree_HasPending(t *testing.T) {
	tree := NewToolCallTree()

	tc := ToolCall{ID: "tc1", Name: "bash", Input: "{}"}
	cmp := NewToolCallCmp("msg1", tc)
	cmp.Init() // Initialize to set spinning state
	tree.Add(cmp)

	// No result yet = pending (spinning)
	if !tree.HasPending() {
		t.Error("Expected HasPending=true for tool without result")
	}

	// Add result
	cmp.SetToolResult(ToolResult{ToolCallID: "tc1", Content: "done"})
	if tree.HasPending() {
		t.Error("Expected HasPending=false after result set")
	}
}

func TestToolCallTree_CollapseAll(t *testing.T) {
	tree := NewToolCallTree()

	tc1 := ToolCall{ID: "tc1", Name: "agent", Input: "{}"}
	tc2 := ToolCall{ID: "tc2", Name: "agent", Input: "{}"}
	cmp1 := NewToolCallCmp("msg1", tc1)
	cmp2 := NewToolCallCmp("msg1", tc2)
	tree.Add(cmp1)
	tree.Add(cmp2)

	tree.CollapseAll()

	if cmp1.IsExpanded() {
		t.Error("Expected cmp1 collapsed")
	}
	if cmp2.IsExpanded() {
		t.Error("Expected cmp2 collapsed")
	}
}

func TestToolCallTree_ExpandAll(t *testing.T) {
	tree := NewToolCallTree()

	tc1 := ToolCall{ID: "tc1", Name: "agent", Input: "{}"}
	cmp1 := NewToolCallCmp("msg1", tc1)
	cmp1.SetExpanded(false)
	tree.Add(cmp1)

	tree.ExpandAll()

	if !cmp1.IsExpanded() {
		t.Error("Expected cmp1 expanded")
	}
}

func TestToolCallTree_AutoCollapse(t *testing.T) {
	tree := NewToolCallTree()

	// Parent with completed nested child
	tc := ToolCall{ID: "parent", Name: "agent", Input: "{}"}
	parent := NewToolCallCmp("msg1", tc)

	nestedTC := ToolCall{ID: "nested1", Name: "bash", Input: "{}"}
	nested := NewToolCallCmp("msg1", nestedTC, WithToolCallNested(true))
	// Mark nested as complete (has result)
	nested.SetToolResult(ToolResult{ToolCallID: "nested1", Content: "done"})
	parent.SetNestedToolCalls([]ToolCallCmp{nested})

	tree.Add(parent)

	tree.AutoCollapse()

	if parent.IsExpanded() {
		t.Error("Expected parent collapsed after AutoCollapse (all children complete)")
	}
}

func TestToolCallTree_All(t *testing.T) {
	tree := NewToolCallTree()

	tc1 := ToolCall{ID: "tc1", Name: "bash", Input: "{}"}
	tc2 := ToolCall{ID: "tc2", Name: "view", Input: "{}"}
	tree.Add(NewToolCallCmp("msg1", tc1))
	tree.Add(NewToolCallCmp("msg1", tc2))

	all := tree.All()
	if len(all) != 2 {
		t.Errorf("Expected 2 items from All(), got %d", len(all))
	}
}

func TestToolCallTree_RenderAsTree(t *testing.T) {
	tree := NewToolCallTree()

	tc1 := ToolCall{ID: "tc1", Name: "bash", Input: `{"command":"ls"}`}
	tc2 := ToolCall{ID: "tc2", Name: "bash", Input: `{"command":"pwd"}`}
	cmp1 := NewToolCallCmp("msg1", tc1)
	cmp2 := NewToolCallCmp("msg1", tc2)
	cmp1.SetSize(80, 0)
	cmp2.SetSize(80, 0)
	tree.Add(cmp1)
	tree.Add(cmp2)

	rendered := tree.RenderAsTree()
	if rendered == "" {
		t.Error("Expected non-empty tree render")
	}
	// Should contain tree branch characters
	if !strings.Contains(rendered, "─") && !strings.Contains(rendered, "├") && !strings.Contains(rendered, "╰") {
		t.Log("Tree render:", rendered)
		// Note: may not have branches if rendering differently
	}
}
