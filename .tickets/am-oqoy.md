---
id: am-oqoy
status: closed
deps: [am-tzub]
links: []
created: 2026-02-18T03:13:47Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Extract ayo-tickets plugin from current ticket system

Extract the current ticket system into the ayo-tickets planner plugin. This becomes the default long-term planner.

## Context
- Current implementation: internal/tickets/
- Target: internal/planners/builtin/tickets/
- This is the reference implementation for long-term planners

## Current System Analysis
The existing ticket system provides:
- Markdown files with YAML frontmatter in .tickets/
- Status workflow (open, in_progress, blocked, closed)
- Dependencies between tickets
- Assignee tracking
- CLI via 'ayo ticket' commands

## Implementation
1. Create internal/planners/builtin/tickets/plugin.go implementing PlannerPlugin
2. Wrap existing internal/tickets/service.go functionality
3. Expose ticket tools (create, start, close, list, etc.)
4. State directory maps to .tickets/ in sandbox
5. Register with DefaultRegistry in init()

## Files to Create
- internal/planners/builtin/tickets/plugin.go
- internal/planners/builtin/tickets/tools.go

## Files to Modify
- internal/tickets/service.go (make path configurable if hardcoded)

## Dependencies
- am-tzub (PlannerRegistry)

## Acceptance
- ayo-tickets plugin implements PlannerPlugin
- Provides ticket tools (create, list, start, close, etc.)
- Uses existing .tickets/ markdown format
- Registered as 'ayo-tickets' in registry


## Notes

**2026-02-18T03:24:38Z**

Split into atomic tickets: am-gp5a, am-92x6, am-eh10, am-uw8z
