// Package squads provides squad management functionality including escalation.
package squads

import (
	"context"
	"fmt"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/tickets"
)

// EscalationType is the ticket type constant for escalations.
// This mirrors tickets.TypeEscalation for convenience.
const EscalationType = tickets.TypeEscalation

// EscalationTag is the tag applied to all escalation tickets.
const EscalationTag = "escalation"

// EscalationParams are the parameters for the escalate tool.
type EscalationParams struct {
	// Reason describes why escalation is needed.
	Reason string `json:"reason" jsonschema:"required,description=Why this issue needs escalation to squad lead"`

	// Context provides additional information about the issue.
	Context string `json:"context,omitempty" jsonschema:"description=Relevant context about the current state and what has been tried"`

	// Priority is the urgency of the escalation (0-4, default 2).
	Priority int `json:"priority,omitempty" jsonschema:"description=Priority level 0-4 where 0 is highest (default: 2)"`
}

// EscalationResult contains the result of an escalation.
type EscalationResult struct {
	// TicketID is the ID of the created escalation ticket.
	TicketID string `json:"ticket_id"`

	// Message describes what happened.
	Message string `json:"message"`
}

// EscalationToolConfig configures the escalate tool for a squad agent.
type EscalationToolConfig struct {
	// SquadName is the name of the squad.
	SquadName string

	// AgentHandle is the handle of the agent creating the escalation.
	AgentHandle string

	// TicketsDir is the path to the squad's tickets directory.
	TicketsDir string
}

// NewEscalateTool creates an escalate tool for squad agents.
// This tool allows agents within a squad to escalate issues to the squad lead
// by creating a special ticket in the squad's planner.
func NewEscalateTool(cfg EscalationToolConfig) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"escalate",
		"Escalate an issue to the squad lead for resolution. Use when you encounter a problem you cannot solve, need decisions outside your scope, or require coordination with other agents.",
		func(ctx context.Context, params EscalationParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Reason == "" {
				return fantasy.NewTextErrorResponse("reason is required; explain why escalation is needed"), nil
			}

			// Validate priority
			priority := params.Priority
			if !tickets.ValidatePriority(priority) {
				priority = tickets.DefaultPriority
			}

			// Create ticket service for squad's tickets directory
			svc := tickets.NewDirectService(cfg.TicketsDir)

			// Build description
			description := fmt.Sprintf("## Reason\n\n%s", params.Reason)
			if params.Context != "" {
				description = fmt.Sprintf("%s\n\n## Context\n\n%s", description, params.Context)
			}
			description = fmt.Sprintf("%s\n\n## Source\n\nEscalated by `%s` in squad `%s`", description, cfg.AgentHandle, cfg.SquadName)

			// Create the escalation ticket
			// Note: "" sessionID since we're using direct service
			ticket, err := svc.Create("", tickets.CreateOptions{
				Title:       fmt.Sprintf("Escalation from %s: %s", cfg.AgentHandle, truncateTitle(params.Reason)),
				Description: description,
				Type:        EscalationType,
				Priority:    priority,
				Assignee:    "squad-lead",
				Tags:        []string{EscalationTag, "from:" + cfg.AgentHandle},
			})
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to create escalation ticket: %v", err)), nil
			}

			result := EscalationResult{
				TicketID: ticket.ID,
				Message:  fmt.Sprintf("Escalation ticket %s created. Squad lead will be notified.", ticket.ID),
			}

			return fantasy.NewTextResponse(result.Message), nil
		},
	)
}

// truncateTitle shortens a reason to be usable as a ticket title.
func truncateTitle(reason string) string {
	const maxLen = 60
	if len(reason) <= maxLen {
		return reason
	}
	return reason[:maxLen-3] + "..."
}

// IsEscalation returns true if the ticket is an escalation ticket.
func IsEscalation(t *tickets.Ticket) bool {
	if t == nil {
		return false
	}
	if t.Type == EscalationType {
		return true
	}
	for _, tag := range t.Tags {
		if tag == EscalationTag {
			return true
		}
	}
	return false
}

// ListEscalations returns all escalation tickets from a service.
func ListEscalations(svc *tickets.Service, sessionID string) ([]*tickets.Ticket, error) {
	all, err := svc.List(sessionID, tickets.Filter{})
	if err != nil {
		return nil, err
	}

	var escalations []*tickets.Ticket
	for _, t := range all {
		if IsEscalation(t) {
			escalations = append(escalations, t)
		}
	}
	return escalations, nil
}
