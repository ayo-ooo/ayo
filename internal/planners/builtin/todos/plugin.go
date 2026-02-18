// Package todos provides the ayo-todos near-term planner plugin.
// This planner manages session-scoped todo items for tracking immediate work.
package todos

import (
	"context"
	"path/filepath"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/planners"
)

// PluginName is the identifier for this planner in the registry.
const PluginName = "ayo-todos"

// Plugin implements the PlannerPlugin interface for ayo-todos.
// It provides session-scoped todo list management for agents.
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

// Type returns the planner type (near-term for todos).
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
// Tool definitions are in tools.go.
func (p *Plugin) Tools() []fantasy.AgentTool {
	return []fantasy.AgentTool{
		p.newTodosTool(),
	}
}

// Instructions returns text to inject into agent system prompts.
// This explains how and when to use the todos tool.
func (p *Plugin) Instructions() string {
	return TodosInstructions
}

// TodosInstructions contains the system prompt instructions for the todos tool.
const TodosInstructions = `## Near-Term Task Management

Use the todos tool to track progress on complex, multi-step tasks:

**When to use:**
- Tasks requiring 3+ distinct steps
- Multi-file changes that need tracking
- Work that benefits from explicit progress tracking

**How to use:**
- Create specific, actionable todo items
- Keep exactly ONE task in_progress at a time
- Mark tasks complete IMMEDIATELY after finishing
- Update todos proactively as work progresses

**Task states:**
- pending: Not yet started
- in_progress: Currently working on (only one)
- completed: Finished successfully

**Required fields for each todo:**
- content: What needs to be done (imperative form, e.g., "Run tests")
- active_form: Present continuous form (e.g., "Running tests")
- status: One of pending, in_progress, completed

**Best practices:**
- Break complex tasks into smaller steps
- Don't batch completions; mark done immediately
- Remove irrelevant tasks from the list
- The user can see your todo list in real-time
`

// StateDir returns the directory where this planner stores its state.
func (p *Plugin) StateDir() string {
	return p.stateDir
}

// State returns the current state for this planner.
// Returns nil if Init has not been called.
func (p *Plugin) State() *State {
	return p.state
}

// statePath returns the full path to the state file.
func (p *Plugin) statePath() string {
	return filepath.Join(p.stateDir, StateFile)
}

// Register adds the ayo-todos plugin to the default registry.
// This is called from init() to ensure the plugin is available at startup.
func Register() {
	planners.DefaultRegistry.Register(PluginName, New())
}

func init() {
	Register()
}
