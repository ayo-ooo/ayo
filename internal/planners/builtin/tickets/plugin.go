// Package tickets provides the ayo-tickets long-term planner plugin.
// This planner manages persistent tickets for tracking work across sessions.
package tickets

import (
	"context"
	"os"

	"charm.land/fantasy"
	"github.com/alexcabrera/ayo/internal/planners"
	"github.com/alexcabrera/ayo/internal/tickets"
)

// PluginName is the identifier for this planner in the registry.
const PluginName = "ayo-tickets"

// Plugin implements the PlannerPlugin interface for ayo-tickets.
// It provides long-term ticket management for agents.
type Plugin struct {
	stateDir string
	service  *tickets.Service
}

// New returns a factory function that creates new Plugin instances.
// This is used by the planner registry to instantiate the planner.
func New() planners.PlannerFactory {
	return func(ctx planners.PlannerContext) (planners.PlannerPlugin, error) {
		// Create state directory if it doesn't exist
		if err := os.MkdirAll(ctx.StateDir, 0o755); err != nil {
			return nil, err
		}

		// Create direct service (tickets stored directly in state dir)
		svc := tickets.NewDirectService(ctx.StateDir)

		return &Plugin{
			stateDir: ctx.StateDir,
			service:  svc,
		}, nil
	}
}

// Name returns the unique identifier for this planner.
func (p *Plugin) Name() string {
	return PluginName
}

// Type returns the planner type (long-term for tickets).
func (p *Plugin) Type() planners.PlannerType {
	return planners.LongTerm
}

// Init initializes the planner.
// For tickets, most initialization happens in the factory since we need the service.
func (p *Plugin) Init(ctx context.Context) error {
	return nil
}

// Close releases any resources held by the planner.
func (p *Plugin) Close() error {
	return nil
}

// Tools returns the fantasy tools that this planner provides.
// Tool definitions will be implemented in am-92x6.
func (p *Plugin) Tools() []fantasy.AgentTool {
	// Tools will be added in am-92x6 (Implement ayo-tickets tool definitions)
	return nil
}

// Instructions returns text to inject into agent system prompts.
// Instructions will be implemented in am-eh10.
func (p *Plugin) Instructions() string {
	// Instructions will be added in am-eh10 (Add ayo-tickets instructions for system prompt)
	return ""
}

// StateDir returns the directory where this planner stores its state.
func (p *Plugin) StateDir() string {
	return p.stateDir
}

// Service returns the underlying ticket service.
// Returns nil if the plugin has not been initialized.
func (p *Plugin) Service() *tickets.Service {
	return p.service
}

// Register adds the ayo-tickets plugin to the default registry.
// This is called from init() to ensure the plugin is available at startup.
func Register() {
	planners.DefaultRegistry.Register(PluginName, New())
}

func init() {
	Register()
}
