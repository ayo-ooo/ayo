---
id: am-e2u1
status: closed
deps: [am-0yb4]
links: []
created: 2026-02-18T03:16:12Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-d01h
---
# Implement escalation via squad planner

Allow agents within a squad to escalate issues to squad lead via planner.

## Context
- Agents in a squad cannot directly call @ayo-main
- Escalation creates a ticket in squad's planner for lead to handle
- Lead may eventually bubble up to @ayo-main via squad output

## Implementation
```go
// Escalation creates a special ticket type

type EscalationTicket struct {
    Type        string // "escalation"
    FromAgent   string
    Reason      string
    Context     map[string]any
}
```

Add escalate tool to squad agents:
```json
{
  "name": "escalate",
  "description": "Escalate an issue to squad lead for resolution",
  "parameters": {
    "reason": "Why this needs escalation",
    "context": "Relevant context"
  }
}
```

## Files to Create
- internal/squads/escalation.go

## Dependencies
- am-0yb4 (squad lead spawning)

## Acceptance
- Agents can call escalate tool
- Ticket created in squad planner
- Squad lead receives escalation

