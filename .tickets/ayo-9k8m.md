---
id: ayo-9k8m
status: closed
deps: []
links: []
created: 2026-02-23T22:15:47Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-xfu3
tags: [squads, lead]
---
# Clarify squad lead agent semantics

Document and implement clear squad lead behavior. The lead agent receives unrouted dispatches, coordinates work through tickets, and synthesizes output from other agents.

## Context

Squad lead is a special role with unique responsibilities:
- Receives all dispatches that aren't routed to specific agents
- Decomposes tasks into tickets for other agents
- Reviews completed work and synthesizes final output
- Does NOT directly edit files (delegates to worker agents)

## Lead Agent Capabilities

| Capability | Lead | Worker |
|------------|------|--------|
| Receive unrouted dispatches | ✓ | ✗ |
| ticket_create | ✓ | ✗ |
| ticket_assign | ✓ | ✗ |
| ticket_delegate | ✓ | ✗ |
| ticket_review | ✓ | ✗ |
| file editing (bash, edit) | ✗ | ✓ |
| ticket_start/close | ✓ | ✓ |

## Tool Configuration

Lead agents automatically get coordination tools:

```go
// internal/squads/squad_lead.go
func GetLeadAgentTools() []string {
    return []string{
        "ticket_create",
        "ticket_assign",
        "ticket_delegate",
        "ticket_review",
        "ticket_list",
        "ticket_close",
        // Note: NO bash, edit, write tools
    }
}
```

## Lead Behavior

### On Receiving Dispatch

```
1. Analyze the request
2. Break into atomic tasks (tickets)
3. Assign tickets to appropriate agents
4. Wait for completion
5. Review and synthesize results
6. Return final response
```

### Ticket Creation Flow

```go
// Lead creates ticket
ticket := &Ticket{
    Title:    "Implement user authentication",
    Assignee: "@backend",
    Priority: 1,
    Body:     "Implement JWT-based auth...",
}

// System notifies @backend
// @backend starts work, commits changes
// @backend closes ticket with summary

// Lead reviews and continues
```

## AGENT.md for Lead

```markdown
# @architect (Squad Lead)

You are the lead architect for this squad. Your role is to:

1. **Decompose** - Break user requests into atomic tasks
2. **Delegate** - Assign tasks to appropriate agents via tickets
3. **Coordinate** - Ensure agents have what they need
4. **Review** - Check completed work meets requirements
5. **Synthesize** - Combine results into coherent response

You do NOT edit files directly. Use tickets to delegate work.

## Available Agents
- @frontend - UI components, React, CSS
- @backend - API endpoints, database, auth
- @qa - Testing, quality assurance
```

## Files to Modify

1. **`internal/squads/squad_lead.go`** - Lead-specific logic
2. **`internal/squads/service.go`** - Identify lead, configure tools
3. **`internal/tools/ticket_delegate.go`** (new) - Delegation tool
4. **`internal/tools/ticket_review.go`** (new) - Review tool
5. **`docs/squads.md`** - Document lead semantics

## Acceptance Criteria

- [ ] Lead receives unrouted dispatches
- [ ] Lead has coordination tools, not editing tools
- [ ] Lead can create and assign tickets
- [ ] Lead can delegate work to other agents
- [ ] Lead receives notification when tickets complete
- [ ] Lead can review completed work
- [ ] Documentation clearly explains lead role

## Testing

- Test dispatch routing to lead
- Test lead tool availability
- Test ticket_delegate workflow
- Test lead review workflow
