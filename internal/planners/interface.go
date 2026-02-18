package planners

import (
	"context"

	"charm.land/fantasy"
)

// PlannerPlugin is the interface that all planner plugins must implement.
// Planners provide work coordination tools and instructions to agents.
type PlannerPlugin interface {
	// Name returns the unique identifier for this planner.
	// This is used for registration and configuration.
	// Examples: "ayo-todos", "ayo-tickets", "custom-kanban"
	Name() string

	// Type returns whether this is a near-term or long-term planner.
	Type() PlannerType

	// Init initializes the planner with the given context.
	// This is called once when the planner is first instantiated for a sandbox.
	// The StateDir in ctx is guaranteed to exist when this is called.
	Init(ctx context.Context) error

	// Close releases any resources held by the planner.
	// This is called when the sandbox is being shut down.
	Close() error

	// Tools returns the fantasy tools that this planner provides.
	// These tools will be added to agents running in the sandbox.
	// Each tool should handle its own state persistence via StateDir.
	Tools() []fantasy.AgentTool

	// Instructions returns text to inject into agent system prompts.
	// This should explain how and when to use the planner's tools.
	// Keep instructions concise and actionable.
	Instructions() string

	// StateDir returns the directory where this planner stores its state.
	// This is the same value passed in PlannerContext during creation.
	StateDir() string
}

// PlannerFactory is a function that creates a new planner instance.
// It receives the context containing sandbox information and configuration.
type PlannerFactory func(ctx PlannerContext) (PlannerPlugin, error)

// Ensure interface compliance at compile time.
var _ PlannerPlugin = (*basePlanner)(nil)

// basePlanner is an unexported type to verify interface completeness.
// It's never instantiated; it just ensures the interface is implementable.
type basePlanner struct{}

func (b *basePlanner) Name() string                   { return "" }
func (b *basePlanner) Type() PlannerType              { return "" }
func (b *basePlanner) Init(ctx context.Context) error { return nil }
func (b *basePlanner) Close() error                   { return nil }
func (b *basePlanner) Tools() []fantasy.AgentTool     { return nil }
func (b *basePlanner) Instructions() string           { return "" }
func (b *basePlanner) StateDir() string               { return "" }
