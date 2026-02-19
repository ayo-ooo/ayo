// Package goals provides the ayo-goals near-term planner plugin.
// This planner focuses on session goals and outcomes rather than tasks.
package goals

import (
	"context"
	"path/filepath"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/planners"
)

// PluginName is the identifier for this planner in the registry.
const PluginName = "ayo-goals"

// Plugin implements the PlannerPlugin interface for ayo-goals.
// It provides goal-oriented session tracking for agents.
type Plugin struct {
	stateDir string
	state    *State
}

// New returns a factory function that creates new Plugin instances.
// This is used by the planner registry to instantiate the planner.
func New() planners.PlannerFactory {
	return func(ctx planners.PlannerContext) (planners.PlannerPlugin, error) {
		return &Plugin{
			stateDir: ctx.StateDir,
		}, nil
	}
}

// Name returns the unique identifier for this planner.
func (p *Plugin) Name() string {
	return PluginName
}

// Type returns the planner type (near-term for goals).
func (p *Plugin) Type() planners.PlannerType {
	return planners.NearTerm
}

// Init initializes the planner, loading any persisted state.
func (p *Plugin) Init(ctx context.Context) error {
	statePath := p.statePath()
	state, err := LoadState(statePath)
	if err != nil {
		return err
	}
	p.state = state
	return nil
}

// Close releases any resources held by the planner.
// Saves state before closing.
func (p *Plugin) Close() error {
	if p.state != nil {
		return p.state.Save(p.statePath())
	}
	return nil
}

// Tools returns the fantasy tools that this planner provides.
func (p *Plugin) Tools() []fantasy.AgentTool {
	return []fantasy.AgentTool{
		p.newGoalsTool(),
	}
}

// Instructions returns text to inject into agent system prompts.
func (p *Plugin) Instructions() string {
	return GoalsInstructions
}

// GoalsInstructions contains the system prompt instructions for the goals tool.
const GoalsInstructions = `## Session Goal Tracking

Use the goals tool to track what you're trying to achieve in this session.

**Philosophy:**
Goals focus on outcomes rather than steps. Instead of "write tests", set a goal like "ensure the feature works correctly". This keeps focus on what matters rather than specific methods.

**When to use:**
- Start of session: Define what success looks like
- During work: Update progress toward goals
- End of work: Mark goals as achieved

**How to use:**
- Set a primary goal for the session
- Optionally add milestones (checkpoints toward the goal)
- Update progress percentage as work advances
- Add notes explaining what was learned or changed

**Goal states:**
- active: Currently working toward this goal
- achieved: Goal was accomplished
- abandoned: Goal was dropped (provide reason in notes)

**Required fields:**
- goal: What you're trying to achieve (outcome-focused)
- status: active, achieved, or abandoned
- progress: 0-100 percentage estimate

**Optional fields:**
- milestones: Checkpoints toward the goal
- notes: Context, learnings, or blockers

**Best practices:**
- Keep goals outcome-focused, not task-focused
- One primary goal at a time
- Update progress honestly
- Add notes explaining decisions or discoveries
`

// StateDir returns the directory where this planner stores its state.
func (p *Plugin) StateDir() string {
	return p.stateDir
}

// State returns the current state for this planner.
func (p *Plugin) State() *State {
	return p.state
}

// statePath returns the full path to the state file.
func (p *Plugin) statePath() string {
	return filepath.Join(p.stateDir, StateFile)
}

// Register adds the ayo-goals plugin to the default registry.
func Register() {
	planners.DefaultRegistry.Register(PluginName, New())
}

func init() {
	Register()
}
