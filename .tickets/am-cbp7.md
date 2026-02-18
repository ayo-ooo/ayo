---
id: am-cbp7
status: open
deps: [am-mrpg, am-tzub]
links: []
created: 2026-02-18T03:14:39Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Load external planner plugins on startup

Load planner plugins from installed plugin directories and register with PlannerRegistry.

## Context
- Plugin loading: internal/plugins/
- Planners register with global Registry

## Implementation
During plugin load:
1. Check manifest for planners field
2. For each planner, load the plugin binary/module
3. Register factory function with DefaultRegistry

## Files to Modify
- internal/plugins/install.go or new internal/plugins/planners.go

## Dependencies
- am-mrpg (Planner manifest fields)
- am-tzub (PlannerRegistry)

## Acceptance
- External planner plugins loaded on startup
- Registered with DefaultRegistry by name
- Errors surfaced if plugin fails to load

