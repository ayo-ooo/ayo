package agent

import (
	"context"
	"strings"
	"testing"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/planners"
)

// mockPlannerPlugin implements planners.PlannerPlugin for testing.
type mockPlannerPlugin struct {
	name         string
	plannerType  planners.PlannerType
	instructions string
}

func (m *mockPlannerPlugin) Name() string                        { return m.name }
func (m *mockPlannerPlugin) Type() planners.PlannerType          { return m.plannerType }
func (m *mockPlannerPlugin) Instructions() string                { return m.instructions }
func (m *mockPlannerPlugin) Tools() []fantasy.AgentTool          { return nil }
func (m *mockPlannerPlugin) StateDir() string                    { return "" }
func (m *mockPlannerPlugin) Init(ctx context.Context) error      { return nil }
func (m *mockPlannerPlugin) Close() error                        { return nil }

func TestInjectPlannerInstructions_NilPlanners(t *testing.T) {
	agent := Agent{
		CombinedSystem: "test system prompt",
	}

	result := InjectPlannerInstructions(agent, nil)
	if result.CombinedSystem != agent.CombinedSystem {
		t.Error("nil planners should not modify system prompt")
	}
}

func TestInjectPlannerInstructions_BothPlanners(t *testing.T) {
	agent := Agent{
		CombinedSystem: "<environment>\ntest env\n</environment>\n\nOriginal prompt",
	}

	sp := &planners.SandboxPlanners{
		NearTerm: &mockPlannerPlugin{
			name:         "test-near",
			plannerType:  planners.NearTerm,
			instructions: "## Near-term instructions\nUse todos.",
		},
		LongTerm: &mockPlannerPlugin{
			name:         "test-long",
			plannerType:  planners.LongTerm,
			instructions: "## Long-term instructions\nUse tickets.",
		},
	}

	result := InjectPlannerInstructions(agent, sp)

	// Should contain both instructions
	if !strings.Contains(result.CombinedSystem, "Near-term instructions") {
		t.Error("should contain near-term instructions")
	}
	if !strings.Contains(result.CombinedSystem, "Long-term instructions") {
		t.Error("should contain long-term instructions")
	}

	// Should contain planner block tags
	if !strings.Contains(result.CombinedSystem, "<planner_instructions>") {
		t.Error("should contain opening planner_instructions tag")
	}
	if !strings.Contains(result.CombinedSystem, "</planner_instructions>") {
		t.Error("should contain closing planner_instructions tag")
	}

	// Should still contain original prompt
	if !strings.Contains(result.CombinedSystem, "Original prompt") {
		t.Error("should preserve original prompt")
	}
}

func TestInjectPlannerInstructions_OnlyNearTerm(t *testing.T) {
	agent := Agent{
		CombinedSystem: "test prompt",
	}

	sp := &planners.SandboxPlanners{
		NearTerm: &mockPlannerPlugin{
			name:         "test-near",
			instructions: "## Near-term only",
		},
	}

	result := InjectPlannerInstructions(agent, sp)

	if !strings.Contains(result.CombinedSystem, "Near-term only") {
		t.Error("should contain near-term instructions")
	}
}

func TestInjectPlannerInstructions_EmptyInstructions(t *testing.T) {
	agent := Agent{
		CombinedSystem: "test prompt",
	}

	sp := &planners.SandboxPlanners{
		NearTerm: &mockPlannerPlugin{
			name:         "test-near",
			instructions: "", // Empty
		},
		LongTerm: &mockPlannerPlugin{
			name:         "test-long",
			instructions: "", // Empty
		},
	}

	result := InjectPlannerInstructions(agent, sp)

	// Should not modify if no instructions
	if result.CombinedSystem != agent.CombinedSystem {
		t.Error("empty instructions should not modify system prompt")
	}
}

func TestBuildPlannerInstructionsBlock(t *testing.T) {
	instructions := []string{
		"## Part One\nContent one.",
		"## Part Two\nContent two.",
	}

	block := BuildPlannerInstructionsBlock(instructions)

	if !strings.HasPrefix(block, "<planner_instructions>") {
		t.Error("block should start with opening tag")
	}
	if !strings.HasSuffix(block, "</planner_instructions>") {
		t.Error("block should end with closing tag")
	}
	if !strings.Contains(block, "Part One") {
		t.Error("block should contain part one")
	}
	if !strings.Contains(block, "Part Two") {
		t.Error("block should contain part two")
	}
}

func TestBuildPlannerInstructionsBlock_Empty(t *testing.T) {
	block := BuildPlannerInstructionsBlock(nil)
	if block != "" {
		t.Errorf("empty instructions should return empty block, got %q", block)
	}

	block = BuildPlannerInstructionsBlock([]string{})
	if block != "" {
		t.Errorf("empty slice should return empty block, got %q", block)
	}
}

func TestInjectPlannerBlock_AfterEnvironment(t *testing.T) {
	systemPrompt := "<environment>\ntest\n</environment>\n\nMain content"
	plannerBlock := "<planner_instructions>\ntest instructions\n</planner_instructions>"

	result := injectPlannerBlock(systemPrompt, plannerBlock)

	// Planner block should be after environment block
	envIdx := strings.Index(result, "</environment>")
	plannerIdx := strings.Index(result, "<planner_instructions>")
	if plannerIdx <= envIdx {
		t.Error("planner block should be after environment block")
	}

	// Main content should still be present and after planner block
	contentIdx := strings.Index(result, "Main content")
	if contentIdx <= plannerIdx {
		t.Error("main content should be after planner block")
	}
}

func TestInjectPlannerBlock_NoEnvironment(t *testing.T) {
	systemPrompt := "Main content only"
	plannerBlock := "<planner_instructions>\ntest instructions\n</planner_instructions>"

	result := injectPlannerBlock(systemPrompt, plannerBlock)

	// Planner block should be prepended
	if !strings.HasPrefix(result, "<planner_instructions>") {
		t.Error("planner block should be prepended when no environment block")
	}

	if !strings.Contains(result, "Main content only") {
		t.Error("should preserve original content")
	}
}

func TestInjectPlannerBlock_Empty(t *testing.T) {
	systemPrompt := "test"
	result := injectPlannerBlock(systemPrompt, "")
	if result != systemPrompt {
		t.Error("empty block should not modify prompt")
	}
}

func TestGetPlannerInstructionsFromPlanners(t *testing.T) {
	sp := &planners.SandboxPlanners{
		NearTerm: &mockPlannerPlugin{
			instructions: "near term instructions",
		},
		LongTerm: &mockPlannerPlugin{
			instructions: "long term instructions",
		},
	}

	pi := GetPlannerInstructionsFromPlanners(sp)

	if pi.NearTerm != "near term instructions" {
		t.Errorf("NearTerm = %q, want %q", pi.NearTerm, "near term instructions")
	}
	if pi.LongTerm != "long term instructions" {
		t.Errorf("LongTerm = %q, want %q", pi.LongTerm, "long term instructions")
	}
}

func TestGetPlannerInstructionsFromPlanners_Nil(t *testing.T) {
	pi := GetPlannerInstructionsFromPlanners(nil)

	if pi.NearTerm != "" || pi.LongTerm != "" {
		t.Error("nil planners should return empty instructions")
	}
}
