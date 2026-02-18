package agent

import (
	"slices"

	"github.com/alexcabrera/ayo/internal/squads"
)

// SquadLeadRestrictedTools is the list of tools that squad leads cannot use.
// These tools would allow the squad lead to reach outside its sandbox boundary.
var SquadLeadRestrictedTools = []string{
	"dispatch_squad",  // Cannot dispatch to other squads
	"invoke_agent",    // Cannot invoke agents outside squad
	"cross_sandbox",   // Cannot access cross-sandbox resources
}

// CreateSquadLead creates an @ayo variant scoped to a specific squad.
// The returned agent:
// - Has the squad constitution injected into its system prompt
// - Cannot delegate outside the squad (delegates cleared)
// - Cannot use tools that reach outside the squad (restricted tools set)
// - Is marked as a squad lead for downstream processing
// - Still appears as @ayo to the user
func CreateSquadLead(baseAyo Agent, constitution *squads.Constitution) (Agent, error) {
	// Copy the base agent
	lead := baseAyo

	// Mark as squad lead
	lead.IsSquadLead = true

	// Get squad name from constitution
	if constitution != nil {
		lead.SquadName = constitution.SquadName
	}

	// Force sandbox enabled for squad leads
	enabled := true
	lead.Config.Sandbox.Enabled = &enabled

	// Inject constitution into system prompt
	lead.CombinedSystem = squads.InjectConstitution(lead.CombinedSystem, constitution)

	// Clear delegates - squad leads cannot delegate outside the squad
	lead.Config.Delegates = nil

	// Set restricted tools - squad leads cannot use these tools
	lead.RestrictedTools = SquadLeadRestrictedTools

	return lead, nil
}

// RestrictToolsForSquadLead applies squad lead tool restrictions to an agent.
// This can be called after CreateSquadLead to add additional restrictions.
func (a *Agent) RestrictToolsForSquadLead(additionalRestrictions ...string) {
	if !a.IsSquadLead {
		return
	}

	// Start with base restrictions
	restricted := make([]string, len(SquadLeadRestrictedTools))
	copy(restricted, SquadLeadRestrictedTools)

	// Add any additional restrictions
	for _, tool := range additionalRestrictions {
		if !slices.Contains(restricted, tool) {
			restricted = append(restricted, tool)
		}
	}

	a.RestrictedTools = restricted
}

// IsToolRestricted returns true if the given tool is restricted for this agent.
// This should be checked before allowing an agent to use a tool.
func (a Agent) IsToolRestricted(toolName string) bool {
	return slices.Contains(a.RestrictedTools, toolName)
}

// FilterRestrictedTools removes restricted tools from a list of tool names.
// Returns a new slice containing only the allowed tools.
func (a Agent) FilterRestrictedTools(tools []string) []string {
	if len(a.RestrictedTools) == 0 {
		return tools
	}

	allowed := make([]string, 0, len(tools))
	for _, tool := range tools {
		if !a.IsToolRestricted(tool) {
			allowed = append(allowed, tool)
		}
	}
	return allowed
}

// IsSquadLead returns true if this agent is a squad lead.
func (a Agent) IsLeadingSquad() bool {
	return a.IsSquadLead
}

// LeadingSquad returns the name of the squad this agent is leading.
// Returns empty string if not a squad lead.
func (a Agent) LeadingSquad() string {
	return a.SquadName
}
