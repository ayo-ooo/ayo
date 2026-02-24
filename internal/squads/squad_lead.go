// Package squads provides lead agent functionality for squad coordination.
package squads

// LeadTools returns the tools available to squad lead agents.
// Lead agents coordinate work through tickets and do NOT directly edit files.
func LeadTools() []string {
	return []string{
		"ticket_create",
		"ticket_assign",
		"ticket_delegate",
		"ticket_review",
		"ticket_list",
		"ticket_start",
		"ticket_close",
		"memory_search",
		"memory_store",
	}
}

// WorkerTools returns the tools available to squad worker agents.
// Worker agents execute tasks and can edit files, but cannot manage tickets.
func WorkerTools() []string {
	return []string{
		"bash",
		"edit",
		"view",
		"write",
		"grep",
		"glob",
		"ls",
		"ticket_start",
		"ticket_close",
		"memory_search",
		"memory_store",
	}
}

// LeadDisabledTools returns tools that should be disabled for lead agents.
// Lead agents delegate file editing to worker agents.
func LeadDisabledTools() []string {
	return []string{
		"bash",
		"edit",
		"write",
		"multiedit",
	}
}

// WorkerDisabledTools returns tools that should be disabled for worker agents.
// Worker agents cannot create or assign tickets.
func WorkerDisabledTools() []string {
	return []string{
		"ticket_create",
		"ticket_assign",
		"ticket_delegate",
		"ticket_review",
	}
}

// IsLeadRole returns true if the given role is a lead role.
func IsLeadRole(role string) bool {
	leadRoles := map[string]bool{
		"lead":      true,
		"architect": true,
		"planner":   true,
		"pm":        true,
		"manager":   true,
	}
	return leadRoles[role]
}

// GetToolsForRole returns the appropriate tools for a squad role.
func GetToolsForRole(role string, isLead bool) []string {
	if isLead || IsLeadRole(role) {
		return LeadTools()
	}
	return WorkerTools()
}

// GetDisabledToolsForRole returns the tools that should be disabled for a role.
func GetDisabledToolsForRole(role string, isLead bool) []string {
	if isLead || IsLeadRole(role) {
		return LeadDisabledTools()
	}
	return WorkerDisabledTools()
}
