// Package squads provides squad management for agent team coordination.
package squads

import (
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/paths"
)

// Squad represents a squad configuration and metadata.
type Squad struct {
	// Name is squad identifier.
	Name string

	// Config is squad configuration.
	Config config.SquadConfig

	// Status is current squad status.
	Status SquadStatus

	// Schemas contains input/output JSON schemas for validation.
	// Nil if no schemas are defined (free-form mode).
	Schemas *SquadSchemas

	// Constitution is squad's SQUAD.md constitution.
	// Loaded during squad initialization.
	Constitution *Constitution

	// LeadReady indicates if squad lead is ready to accept input.
	// Set to true when squad has been fully initialized.
	LeadReady bool

	// Invoker is used to invoke agents within squad context.
	// If nil, dispatch returns routing info only without actual invocation.
	Invoker AgentInvoker
}

// CanAcceptInput returns true if squad is ready to accept input.
// The squad must be running and have its lead ready.
func (sq *Squad) CanAcceptInput() bool {
	return sq.Status == SquadStatusRunning && sq.LeadReady
}

// GetAllAgents returns all agents in this squad.
// This includes agents from config, constitution, and lead.
func (sq *Squad) GetAllAgents() []string {
	agents := make(map[string]bool)

	// Add agents from config
	for _, a := range sq.Config.Agents {
		agent := a
		if len(agent) > 0 && agent[0] != '@' {
			agent = "@" + agent
		}
		agents[agent] = true
	}

	// Add lead
	lead := sq.Config.Lead
	if lead == "" && sq.Constitution != nil {
		lead = sq.Constitution.Frontmatter.Lead
	}
	if lead == "" {
		lead = "@ayo"
	}
	if len(lead) > 0 && lead[0] != '@' {
		lead = "@" + lead
	}
	agents[lead] = true

	// Add agents from constitution
	if sq.Constitution != nil {
		for _, a := range sq.Constitution.GetAgents() {
			agent := a
			if len(agent) > 0 && agent[0] != '@' {
				agent = "@" + agent
			}
			agents[agent] = true
		}
	}

	// Convert to slice
	result := make([]string, 0, len(agents))
	for a := range agents {
		result = append(result, a)
	}
	return result
}

// HasAgent returns true if given agent is part of this squad.
func (sq *Squad) HasAgent(agentHandle string) bool {
	agent := agentHandle
	if len(agent) > 0 && agent[0] != '@' {
		agent = "@" + agent
	}

	for _, a := range sq.GetAllAgents() {
		if a == agent {
			return true
		}
	}
	return false
}

// IsRunning returns true if squad is currently running.
func (sq *Squad) IsRunning() bool {
	return sq.Status == SquadStatusRunning
}

// SquadStatus represents the status of a squad.
type SquadStatus string

const (
	SquadStatusUnknown  SquadStatus = ""
	SquadStatusStopped  SquadStatus = "stopped"
	SquadStatusRunning  SquadStatus = "running"
	SquadStatusCreating SquadStatus = "creating"
	SquadStatusFailed   SquadStatus = "failed"
)

// GetTicketsDir returns the tickets directory for a squad.
func GetTicketsDir(name string) string {
	return paths.SquadTicketsDir(name)
}

// GetContextDir returns the context directory for a squad.
func GetContextDir(name string) string {
	return paths.SquadContextDir(name)
}

// GetWorkspaceDir returns the workspace directory for a squad.
func GetWorkspaceDir(name string) string {
	return paths.SquadWorkspaceDir(name)
}

// GetTeamWorkspaceDir returns the workspace directory for a team project.
func GetTeamWorkspaceDir(teamDir string) string {
	return paths.TeamWorkspaceDir(teamDir)
}

// GetTeamAgentsDir returns the agents directory for a team project.
func GetTeamAgentsDir(teamDir string) string {
	return paths.TeamAgentsDir(teamDir)
}

// LoadTeamFromProject loads a team configuration from a project directory.
// This is the new team project format for the build system.
func LoadTeamFromProject(teamDir string) (*TeamConfig, error) {
	return LoadTeamConfigFromTOML(teamDir)
}
