package agent

import (
	"strings"

	"github.com/alexcabrera/ayo/internal/planners"
)

// PlannerInstructionsPrompt is a field that can be set on an agent to hold planner instructions.
// This is separate from CombinedSystem to allow for lazy injection.
type PlannerInstructions struct {
	NearTerm string
	LongTerm string
}

// InjectPlannerInstructions modifies the agent's system prompt to include planner instructions.
// This should be called after planners are resolved for the sandbox.
// Returns a copy of the agent with updated CombinedSystem.
func InjectPlannerInstructions(agent Agent, planners *planners.SandboxPlanners) Agent {
	if planners == nil {
		return agent
	}

	// Collect instructions from planners
	var instructions []string

	if planners.NearTerm != nil {
		if instr := planners.NearTerm.Instructions(); instr != "" {
			instructions = append(instructions, instr)
		}
	}

	if planners.LongTerm != nil {
		if instr := planners.LongTerm.Instructions(); instr != "" {
			instructions = append(instructions, instr)
		}
	}

	if len(instructions) == 0 {
		return agent
	}

	// Build combined planner instructions block
	plannerBlock := BuildPlannerInstructionsBlock(instructions)

	// Inject into system prompt
	agent.CombinedSystem = injectPlannerBlock(agent.CombinedSystem, plannerBlock)

	return agent
}

// BuildPlannerInstructionsBlock creates the planner instructions section.
func BuildPlannerInstructionsBlock(instructions []string) string {
	if len(instructions) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "<planner_instructions>")
	parts = append(parts, instructions...)
	parts = append(parts, "</planner_instructions>")

	return strings.Join(parts, "\n\n")
}

// injectPlannerBlock adds the planner instructions block to the system prompt.
// It places the block after the environment section but before the main content.
func injectPlannerBlock(systemPrompt, plannerBlock string) string {
	if plannerBlock == "" {
		return systemPrompt
	}

	// Look for </environment> tag to insert after
	envEndTag := "</environment>"
	if idx := strings.Index(systemPrompt, envEndTag); idx != -1 {
		insertPos := idx + len(envEndTag)
		before := systemPrompt[:insertPos]
		after := systemPrompt[insertPos:]
		return before + "\n\n" + plannerBlock + after
	}

	// If no environment block, prepend to system prompt
	return plannerBlock + "\n\n" + systemPrompt
}

// GetPlannerInstructionsFromPlanners extracts instructions from sandbox planners.
func GetPlannerInstructionsFromPlanners(sandboxPlanners *planners.SandboxPlanners) PlannerInstructions {
	var pi PlannerInstructions

	if sandboxPlanners == nil {
		return pi
	}

	if sandboxPlanners.NearTerm != nil {
		pi.NearTerm = sandboxPlanners.NearTerm.Instructions()
	}

	if sandboxPlanners.LongTerm != nil {
		pi.LongTerm = sandboxPlanners.LongTerm.Instructions()
	}

	return pi
}
