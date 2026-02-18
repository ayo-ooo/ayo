package agent

import (
	"github.com/alexcabrera/ayo/internal/squads"
)

// CreateSquadLead creates an @ayo variant scoped to a specific squad.
// The returned agent:
// - Has the squad constitution injected into its system prompt
// - Cannot delegate outside the squad (delegates cleared)
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

	return lead, nil
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
