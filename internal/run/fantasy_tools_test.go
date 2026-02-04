package run

import (
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

func TestNewFantasyToolSetWithOptions_TodoAlwaysAvailable(t *testing.T) {
	// Test that todo is available by default, even with empty allowed list
	ts := NewFantasyToolSetWithOptions(nil, "", nil, 0, false)
	defer ts.Close()

	// Should have at least todo and bash
	tools := ts.Tools()
	hasTodo := false
	hasBash := false
	for _, tool := range tools {
		info := tool.Info()
		if info.Name == "todo" {
			hasTodo = true
		}
		if info.Name == "bash" {
			hasBash = true
		}
	}

	if !hasTodo {
		t.Error("expected todo tool to be always available by default")
	}
	if !hasBash {
		t.Error("expected bash tool to be available by default")
	}
}

func TestNewFantasyToolSetWithOptions_TodoDisabled(t *testing.T) {
	// Test that todo is not available when disableTodo=true
	ts := NewFantasyToolSetWithOptions([]string{"bash"}, "", nil, 0, true)
	defer ts.Close()

	tools := ts.Tools()
	for _, tool := range tools {
		info := tool.Info()
		if info.Name == "todo" {
			t.Error("expected todo tool to be disabled when disableTodo=true")
		}
	}
}

func TestNewFantasyToolSetWithOptions_TodoNotDuplicated(t *testing.T) {
	// Test that explicitly listing "todo" in allowed doesn't duplicate it
	ts := NewFantasyToolSetWithOptions([]string{"bash", "todo"}, "", nil, 0, false)
	defer ts.Close()

	tools := ts.Tools()
	todoCount := 0
	for _, tool := range tools {
		info := tool.Info()
		if info.Name == "todo" {
			todoCount++
		}
	}

	if todoCount != 1 {
		t.Errorf("expected exactly 1 todo tool, got %d", todoCount)
	}
}

func TestNewFantasyToolSetWithOptions_OldPlanningNameIgnored(t *testing.T) {
	// Test that "planning" (old category name) is ignored and doesn't cause errors
	ts := NewFantasyToolSetWithOptions([]string{"bash", "planning"}, "", nil, 0, false)
	defer ts.Close()

	tools := ts.Tools()
	hasPlanning := false
	for _, tool := range tools {
		info := tool.Info()
		if info.Name == "planning" {
			hasPlanning = true
		}
	}

	if hasPlanning {
		t.Error("'planning' should be ignored (old category name)")
	}
}

func TestNewFantasyToolSetWithOptions_PlanCategoryResolution(t *testing.T) {
	// Test that "plan" category resolves correctly:
	// - If a default is configured (via plugin or ayo.json), it loads that tool
	// - If no default is configured, no tool is loaded for "plan"
	// - The literal string "plan" should never appear as a tool name
	ts := NewFantasyToolSetWithOptions([]string{"bash", "plan"}, "", nil, 0, false)
	defer ts.Close()

	tools := ts.Tools()
	for _, tool := range tools {
		info := tool.Info()
		// "plan" is a category, not a tool - it should never appear as a tool name
		if info.Name == "plan" {
			t.Errorf("category name 'plan' should not appear as a tool name")
		}
	}
}

func TestNewFantasyToolSetWithOptions_StatefulToolsTracked(t *testing.T) {
	// Test that stateful tools (todo) are tracked for cleanup
	ts := NewFantasyToolSetWithOptions(nil, "", nil, 0, false)

	// Should have at least one stateful tool (todo)
	if len(ts.statefulTools) == 0 {
		t.Error("expected stateful tools to be tracked for cleanup")
	}

	// Close should not error
	if err := ts.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestNewFantasyToolSet_WithSandboxExecutor(t *testing.T) {
	// Test that sandbox executor is used when provided
	provider := sandbox.NewNoneProvider()
	sb, err := provider.Create(nil, providers.SandboxCreateOptions{})
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}
	defer provider.Delete(nil, sb.ID, true)

	executor := sandbox.NewExecutor(provider, sb.ID, t.TempDir())

	ts := NewFantasyToolSet(ToolSetOptions{
		AllowedTools:    []string{"bash"},
		SandboxExecutor: executor,
		DisableTodo:     true, // Simplify test
	})
	defer ts.Close()

	// Should have bash tool
	tools := ts.Tools()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool (bash), got %d", len(tools))
	}

	info := tools[0].Info()
	if info.Name != "bash" {
		t.Errorf("expected bash tool, got %s", info.Name)
	}

	// Verify sandbox executor is stored
	if ts.sandboxExecutor == nil {
		t.Error("expected sandbox executor to be stored")
	}
}

func TestNewFantasyToolSet_WithoutSandboxExecutor(t *testing.T) {
	// Test that local bash is used when no sandbox executor provided
	ts := NewFantasyToolSet(ToolSetOptions{
		AllowedTools: []string{"bash"},
		DisableTodo:  true,
	})
	defer ts.Close()

	// Should have bash tool
	tools := ts.Tools()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool (bash), got %d", len(tools))
	}

	info := tools[0].Info()
	if info.Name != "bash" {
		t.Errorf("expected bash tool, got %s", info.Name)
	}

	// Verify no sandbox executor
	if ts.sandboxExecutor != nil {
		t.Error("expected no sandbox executor")
	}
}
