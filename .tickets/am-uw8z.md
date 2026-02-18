---
id: am-uw8z
status: closed
deps: []
links: []
created: 2026-02-18T03:24:34Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Make tickets.Service path configurable

Update tickets.Service to accept configurable state directory.

## Context
- Currently tickets.Service may have hardcoded paths
- Need to accept StateDir from planner context

## Implementation
Check internal/tickets/service.go:
- Ensure NewService accepts dir parameter
- Remove any hardcoded path assumptions

## Files to Modify
- internal/tickets/service.go

## Acceptance
- Service accepts directory parameter
- Works with any valid directory path

