---
id: am-eh10
status: closed
deps: [am-92x6]
links: []
created: 2026-02-18T03:24:29Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Add ayo-tickets instructions for system prompt

Add Instructions() method to ayo-tickets plugin.

## Context
- Instructions tell the agent how to use ticket tools
- Injected into system prompt

## Implementation
```go
// internal/planners/builtin/tickets/plugin.go

func (p *Plugin) Instructions() string {
    return `## Long-Term Planning

Use ticket tools for persistent work tracking:
- ticket_create: Create new work items
- ticket_start: Begin working on a ticket
- ticket_close: Mark ticket as complete
- ticket_block: Mark ticket as blocked
- ticket_note: Add progress notes

Tickets persist across sessions and can have dependencies.
`
}
```

## Files to Modify
- internal/planners/builtin/tickets/plugin.go

## Dependencies
- am-92x6 (tools implementation)

## Acceptance
- Instructions() returns useful guidance
- Documents all ticket operations

