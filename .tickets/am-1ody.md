---
id: am-1ody
status: open
deps: [am-5f4q]
links: []
created: 2026-02-18T03:15:58Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-d01h
---
# Restrict @ayo-in-squad tool access

Ensure squad lead cannot use tools that reach outside the squad sandbox.

## Context
- Squad lead should only see squad-internal resources
- Cannot invoke cross-squad operations
- Cannot use @ayo-main's planner

## Implementation
Remove/disable tools for squad lead:
- Remove delegate tools that target other squads
- Remove cross-sandbox file access
- Limit to squad's planner only

```go
// internal/agent/squad_lead.go

func (a *Agent) restrictToolsForSquadLead() {
    restricted := []string{
        "dispatch_squad",  // Cannot dispatch to other squads
        "invoke_agent",    // Cannot invoke agents outside squad
    }
    a.Config.AllowedTools = filterTools(a.Config.AllowedTools, restricted)
}
```

## Files to Modify
- internal/agent/squad_lead.go

## Dependencies
- am-5f4q (squad lead creation)

## Acceptance
- Squad lead cannot dispatch to other squads
- Squad lead cannot invoke agents outside squad
- Squad lead uses only squad's planners

