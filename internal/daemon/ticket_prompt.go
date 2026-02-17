package daemon

import (
	"fmt"
	"strings"

	"github.com/alexcabrera/ayo/internal/tickets"
)

// BuildTicketPrompt creates the initial prompt for an agent working on a ticket.
func BuildTicketPrompt(ticket *tickets.Ticket, sessionID string) string {
	var sb strings.Builder

	sb.WriteString("# Ticket Assignment\n\n")
	sb.WriteString("You have been assigned to work on the following ticket.\n\n")

	// Ticket metadata
	sb.WriteString(fmt.Sprintf("**Ticket ID:** %s\n", ticket.ID))
	sb.WriteString(fmt.Sprintf("**Title:** %s\n", ticket.Title))
	sb.WriteString(fmt.Sprintf("**Priority:** P%d\n", ticket.Priority))
	if len(ticket.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(ticket.Tags, ", ")))
	}
	sb.WriteString("\n")

	// Description
	if ticket.Description != "" {
		sb.WriteString("## Description\n\n")
		sb.WriteString(ticket.Description)
		sb.WriteString("\n\n")
	}

	// Dependencies
	if len(ticket.Deps) > 0 {
		sb.WriteString("## Dependencies\n\n")
		sb.WriteString("This ticket depends on the following completed tickets:\n")
		for _, dep := range ticket.Deps {
			sb.WriteString(fmt.Sprintf("- %s\n", dep))
		}
		sb.WriteString("\n")
	}

	// Instructions
	sb.WriteString("## Instructions\n\n")
	sb.WriteString("Work autonomously to complete this ticket. When finished:\n")
	sb.WriteString("1. Verify all requirements are met\n")
	sb.WriteString("2. Run relevant tests\n")
	sb.WriteString("3. Use `ayo ticket close " + ticket.ID + "` to close the ticket\n\n")

	sb.WriteString("If blocked, use `ayo ticket update " + ticket.ID + " --status blocked` ")
	sb.WriteString("and explain the blocker.\n")

	return sb.String()
}
