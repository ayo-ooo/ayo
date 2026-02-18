---
id: am-fl7k
status: closed
deps: [am-tzub, am-1gwl]
links: []
created: 2026-02-18T03:20:23Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Add ayo planner commands

Add CLI commands for managing planners.

## Commands
```bash
ayo planner list              # List available planners
ayo planner show <name>       # Show planner details
ayo planner set near <name>   # Set default near-term planner
ayo planner set long <name>   # Set default long-term planner
```

## Files to Create
- cmd/ayo/planner.go

## Dependencies
- am-tzub (PlannerRegistry)
- am-1gwl (PlannersConfig)

## Acceptance
- Commands work with installed planners
- Config updates persisted
- Clear output for list/show

