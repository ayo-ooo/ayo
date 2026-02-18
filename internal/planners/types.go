package planners

// PlannerType indicates whether a planner is for near-term or long-term planning.
type PlannerType string

const (
	// NearTerm planners handle session-scoped work tracking.
	// Examples: todo lists, immediate task queues.
	// State may be ephemeral or persisted across sessions.
	NearTerm PlannerType = "near"

	// LongTerm planners handle persistent work coordination.
	// Examples: ticket systems, kanban boards, issue trackers.
	// State always persists across sessions.
	LongTerm PlannerType = "long"
)

// String returns the string representation of the planner type.
func (t PlannerType) String() string {
	return string(t)
}

// IsValid returns true if the planner type is a known value.
func (t PlannerType) IsValid() bool {
	return t == NearTerm || t == LongTerm
}

// PlannerContext provides configuration and paths for planner initialization.
// This is passed to the planner factory when creating a new instance.
type PlannerContext struct {
	// SandboxName is the name of the sandbox this planner belongs to.
	// For squads, this is the squad name. For @ayo, this is "ayo".
	SandboxName string

	// SandboxDir is the root directory of the sandbox.
	// Example: ~/.local/share/ayo/sandboxes/squads/frontend-team/
	SandboxDir string

	// StateDir is where the planner should store its state.
	// This directory is created before Init() is called.
	// Example: ~/.local/share/ayo/sandboxes/squads/frontend-team/.planner.near/
	StateDir string

	// Config contains planner-specific configuration from the user.
	// This comes from config.toml or SQUAD.md frontmatter.
	Config map[string]any
}

// PlannersConfig holds the planner selections for a sandbox.
type PlannersConfig struct {
	// NearTerm is the name of the near-term planner to use.
	// Default: "ayo-todos"
	NearTerm string `json:"near_term,omitempty" toml:"near_term" yaml:"near_term"`

	// LongTerm is the name of the long-term planner to use.
	// Default: "ayo-tickets"
	LongTerm string `json:"long_term,omitempty" toml:"long_term" yaml:"long_term"`
}

// WithDefaults returns a copy of the config with default values applied.
func (c PlannersConfig) WithDefaults() PlannersConfig {
	result := c
	if result.NearTerm == "" {
		result.NearTerm = "ayo-todos"
	}
	if result.LongTerm == "" {
		result.LongTerm = "ayo-tickets"
	}
	return result
}

// IsEmpty returns true if no planners are configured.
func (c PlannersConfig) IsEmpty() bool {
	return c.NearTerm == "" && c.LongTerm == ""
}
