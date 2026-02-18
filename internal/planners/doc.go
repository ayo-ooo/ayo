// Package planners provides the plugin system for near-term and long-term planning tools.
//
// Planners are pluggable modules that provide work coordination capabilities to agents.
// Each sandbox (squad or @ayo) can have its own planner instances for tracking work.
//
// There are two types of planners:
//
//   - NearTerm: Session-scoped work tracking (e.g., todos)
//   - LongTerm: Persistent work coordination (e.g., tickets)
//
// Planners expose tools to agents via the fantasy.AgentTool interface and inject
// instructions into system prompts to guide agents on how to use them.
//
// # Built-in Planners
//
// The following planners ship with ayo:
//
//   - ayo-todos: Near-term planning via a simple todo list
//   - ayo-tickets: Long-term planning via markdown tickets with dependencies
//
// # Creating Custom Planners
//
// To create a custom planner, implement the [PlannerPlugin] interface and register
// it with the [DefaultRegistry]:
//
//	func init() {
//	    planners.DefaultRegistry.Register("my-planner", func(ctx planners.PlannerContext) (planners.PlannerPlugin, error) {
//	        return &MyPlanner{stateDir: ctx.StateDir}, nil
//	    })
//	}
//
// # Configuration
//
// Default planners are configured in ~/.config/ayo/config.toml:
//
//	[planners]
//	near_term = "ayo-todos"
//	long_term = "ayo-tickets"
//
// Squads can override defaults in SQUAD.md frontmatter:
//
//	---
//	planners:
//	  near_term: ayo-todos
//	  long_term: custom-kanban
//	---
package planners
