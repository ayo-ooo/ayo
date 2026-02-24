---
id: ayo-akqw
status: open
deps: []
links: []
created: 2026-02-23T22:15:54Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-xfu3
tags: [squads, tickets]
---
# Polish ticket tools for squad agents

Ensure all ticket tools work correctly for agents inside squads. Make tools intuitive and add filtering capabilities.

## Context

Squad agents coordinate via tickets. This ticket polishes the ticket tools to work seamlessly in squad context.

## Ticket Tools

### ticket_create

Create a new ticket:

```json
{
  "name": "ticket_create",
  "arguments": {
    "title": "Implement login page",
    "priority": 1,
    "assignee": "@frontend",
    "body": "Build the login page with...",
    "labels": ["feature", "auth"]
  }
}
```

### ticket_list

List tickets with filtering:

```json
{
  "name": "ticket_list",
  "arguments": {
    "assignee": "@me",        // Filter by assignee
    "status": "open",         // open, in_progress, closed
    "priority": 1,            // Filter by priority
    "labels": ["auth"]        // Filter by labels
  }
}
```

Output:
```
ID        TITLE                    STATUS      ASSIGNEE    PRIORITY
ayo-1234  Implement login page     in_progress @frontend   P1
ayo-1235  Add password reset       open        @frontend   P2
ayo-1236  Write auth tests         open        @qa         P2
```

### ticket_start

Start working on a ticket:

```json
{
  "name": "ticket_start",
  "arguments": {
    "id": "ayo-1234"
  }
}
```

### ticket_close

Close a ticket with summary:

```json
{
  "name": "ticket_close",
  "arguments": {
    "id": "ayo-1234",
    "summary": "Implemented login page with email/password form",
    "status": "done"  // done, wontfix, duplicate
  }
}
```

### ticket_assign

Reassign a ticket:

```json
{
  "name": "ticket_assign",
  "arguments": {
    "id": "ayo-1234",
    "assignee": "@backend"
  }
}
```

## @me Syntax

`@me` resolves to the current agent's identity:

```go
func (t *TicketTools) resolveAssignee(assignee string) string {
    if assignee == "@me" {
        return t.currentAgent
    }
    return assignee
}
```

## Tool Availability

| Tool | Lead | Worker |
|------|------|--------|
| ticket_create | ✓ | ✗ (request via lead) |
| ticket_list | ✓ | ✓ |
| ticket_start | ✓ | ✓ |
| ticket_close | ✓ | ✓ |
| ticket_assign | ✓ | ✗ |

## Files to Modify

1. **`internal/tools/ticket_create.go`** - Polish creation
2. **`internal/tools/ticket_list.go`** - Add filters
3. **`internal/tools/ticket_start.go`** - Polish start
4. **`internal/tools/ticket_close.go`** - Add summary
5. **`internal/tools/ticket_assign.go`** - Polish assignment
6. **`internal/tickets/tickets.go`** - Query support for filters

## Acceptance Criteria

- [ ] ticket_create works for lead agents
- [ ] ticket_list supports all filters
- [ ] @me syntax resolves correctly
- [ ] ticket_start transitions to in_progress
- [ ] ticket_close requires summary
- [ ] ticket_assign updates assignee
- [ ] Workers can't create/assign (only lead)
- [ ] Clear error messages for permission issues

## Testing

- Test each tool in squad context
- Test @me resolution
- Test filter combinations
- Test permission enforcement
