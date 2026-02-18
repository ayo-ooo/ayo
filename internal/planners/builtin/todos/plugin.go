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
// Instructions will be implemented in am-rozh.
func (p *Plugin) Instructions() string {
	// Instructions will be added in am-rozh (Add ayo-todos instructions for system prompt)
	return ""
}

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
