package run

import (
	"context"
	"testing"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

// NewMockPlannerTool creates a simple mock tool for testing planner tool injection.
func NewMockPlannerTool(name, description string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		name,
		description,
		func(ctx context.Context, params struct{}, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return fantasy.NewTextResponse("mock response"), nil
		},
	)
}

// mockPlannerWithTools implements the interface needed for GetPlannerTools.
type mockPlannerWithTools struct {
	tools []fantasy.AgentTool
}

func (m *mockPlannerWithTools) Tools() []fantasy.AgentTool {
	return m.tools
}

func TestNewFantasyToolSet_DefaultBash(t *testing.T) {
	// Test that bash is available by default with empty allowed list
	ts := NewFantasyToolSet(ToolSetOptions{})
	defer ts.Close()

	// Should have bash by default
	tools := ts.Tools()
	hasBash := false
	for _, tool := range tools {
		info := tool.Info()
		if info.Name == "bash" {
			hasBash = true
		}
	}

	if !hasBash {
		t.Error("expected bash tool to be available by default")
	}
}

func TestNewFantasyToolSetWithOptions_PlannerToolsProvided(t *testing.T) {
	// Test that planner tools are included when provided via PlannerTools option
	mockTodosTool := NewMockPlannerTool("todos", "Mock todos tool from planner")

	ts := NewFantasyToolSet(ToolSetOptions{
		AllowedTools: []string{"bash"},
		PlannerTools: []fantasy.AgentTool{mockTodosTool},
	})
	defer ts.Close()

	tools := ts.Tools()
	hasTodos := false
	for _, tool := range tools {
		info := tool.Info()
		if info.Name == "todos" {
			hasTodos = true
		}
	}

	if !hasTodos {
		t.Error("expected todos tool from planners to be available")
	}
}

func TestNewFantasyToolSet_LegacyTodoIgnored(t *testing.T) {
	// Test that explicitly listing "todo" or "todos" in allowed is ignored
	// (these are now provided by planners, not built-in)
	ts := NewFantasyToolSet(ToolSetOptions{AllowedTools: []string{"bash", "todo", "todos"}})
	defer ts.Close()

	tools := ts.Tools()
	for _, tool := range tools {
		info := tool.Info()
		if info.Name == "todo" || info.Name == "todos" {
			t.Error("legacy todo/todos names should be ignored in allowed list")
		}
	}
}

func TestNewFantasyToolSet_OldPlanningNameIgnored(t *testing.T) {
	// Test that "planning" (old category name) is ignored and doesn't cause errors
	ts := NewFantasyToolSet(ToolSetOptions{AllowedTools: []string{"bash", "planning"}})
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

func TestNewFantasyToolSet_PlanCategoryResolution(t *testing.T) {
	// Test that "plan" category resolves correctly:
	// - If a default is configured (via plugin or ayo.json), it loads that tool
	// - If no default is configured, no tool is loaded for "plan"
	// - The literal string "plan" should never appear as a tool name
	ts := NewFantasyToolSet(ToolSetOptions{AllowedTools: []string{"bash", "plan"}})
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

func TestNewFantasyToolSet_StatefulToolsTracked(t *testing.T) {
	// Test that stateful tools are tracked for cleanup when planner tools are provided
	// Without planner tools, there are no stateful tools (the built-in todo was removed)
	ts := NewFantasyToolSet(ToolSetOptions{})

	// Without planner tools, statefulTools should be empty
	if len(ts.statefulTools) != 0 {
		t.Errorf("expected no stateful tools without planner injection, got %d", len(ts.statefulTools))
	}

	// Close should not error even with no stateful tools
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

	executor := sandbox.NewExecutor(provider, sb.ID, t.TempDir(), "")

	ts := NewFantasyToolSet(ToolSetOptions{
		AllowedTools:    []string{"bash"},
		SandboxExecutor: executor,
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

func TestNewFantasyToolSet_WithPlannerTools(t *testing.T) {
	// Create a mock planner tool
	mockTool := NewMockPlannerTool("test_planner_tool", "Test planner tool")

	ts := NewFantasyToolSet(ToolSetOptions{
		AllowedTools: []string{"bash"},
		PlannerTools: []fantasy.AgentTool{mockTool},
	})
	defer ts.Close()

	// Should have bash + planner tool
	tools := ts.Tools()
	if len(tools) != 2 {
		t.Errorf("expected 2 tools (bash + planner), got %d", len(tools))
	}

	hasPlannerTool := false
	for _, tool := range tools {
		if tool.Info().Name == "test_planner_tool" {
			hasPlannerTool = true
		}
	}
	if !hasPlannerTool {
		t.Error("expected planner tool to be included")
	}
}

func TestNewFantasyToolSet_PlannerToolsNoCollision(t *testing.T) {
	// Test that planner tools don't duplicate existing tools
	mockTool := NewMockPlannerTool("bash", "This should not override bash")

	ts := NewFantasyToolSet(ToolSetOptions{
		AllowedTools: []string{"bash"},
		PlannerTools: []fantasy.AgentTool{mockTool},
	})
	defer ts.Close()

	// Should still have only 1 bash tool (not duplicated)
	tools := ts.Tools()
	bashCount := 0
	for _, tool := range tools {
		if tool.Info().Name == "bash" {
			bashCount++
		}
	}
	if bashCount != 1 {
		t.Errorf("expected 1 bash tool (no collision), got %d", bashCount)
	}
}

func TestGetPlannerTools(t *testing.T) {
	// Test helper with both planners
	nearTerm := &mockPlannerWithTools{tools: []fantasy.AgentTool{
		NewMockPlannerTool("near_tool", "Near term tool"),
	}}
	longTerm := &mockPlannerWithTools{tools: []fantasy.AgentTool{
		NewMockPlannerTool("long_tool", "Long term tool"),
	}}

	tools := GetPlannerTools(nearTerm, longTerm)
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
}

func TestGetPlannerTools_NilPlanners(t *testing.T) {
	tools := GetPlannerTools(nil, nil)
	if len(tools) != 0 {
		t.Errorf("expected 0 tools for nil planners, got %d", len(tools))
	}
}

func TestGetPlannerTools_PartialPlanners(t *testing.T) {
	nearTerm := &mockPlannerWithTools{tools: []fantasy.AgentTool{
		NewMockPlannerTool("near_tool", "Near term tool"),
	}}

	tools := GetPlannerTools(nearTerm, nil)
	if len(tools) != 1 {
		t.Errorf("expected 1 tool (near-term only), got %d", len(tools))
	}
}
