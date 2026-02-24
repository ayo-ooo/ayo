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
func (p *Plugin) Tools() []fantasy.AgentTool {
	return []fantasy.AgentTool{
		p.newCreateTool(),
		p.newListTool(),
		p.newStartTool(),
		p.newCloseTool(),
		p.newBlockTool(),
		p.newNoteTool(),
		p.newAssignTool(),
	}
}

// Instructions returns text to inject into agent system prompts.
// This explains how and when to use the ticket tools.
func (p *Plugin) Instructions() string {
	return TicketsInstructions
}

// TicketsInstructions contains the system prompt instructions for the ticket tools.
const TicketsInstructions = `## Long-Term Work Planning

Use ticket tools for persistent work tracking across sessions:

**When to use:**
- Work items that may span multiple sessions
- Tasks with dependencies on other work
- Tracking larger initiatives or epics
- Coordinating work with other agents

**Available tools:**
- ticket_create: Create new work items with title, description, type, priority
- ticket_list: List tickets with optional filters (status, type, assignee, priority)
- ticket_start: Begin working on a ticket (sets status to in_progress)
- ticket_close: Mark a ticket as complete (with optional closing message)
- ticket_block: Mark a ticket as blocked
- ticket_note: Add timestamped progress notes to a ticket
- ticket_assign: Assign or reassign a ticket to an agent

**Ticket states:**
- open: Ready to be worked on
- in_progress: Currently being worked on
- blocked: Cannot proceed (dependency, external, etc.)
- closed: Work completed

**Ticket types:**
- task: Standard unit of work (default)
- feature: New functionality
- bug: Defect to fix
- chore: Maintenance work
- epic: Large initiative containing sub-tickets

**Priority levels:**
- 0: Critical/urgent
- 1: High priority
- 2: Normal (default)
- 3: Low priority
- 4: Nice to have

**Best practices:**
- Create tickets for work that won't be completed immediately
- Use dependencies to track blockers
- Add notes as you make progress
- Close tickets with a summary message
- Use ticket_list to see what's ready to work on
- Assign tickets to coordinate work in squads
`

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
