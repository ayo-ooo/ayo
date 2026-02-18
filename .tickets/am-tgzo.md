---
id: am-tgzo
status: closed
deps: [am-tzub]
links: []
created: 2026-02-18T03:13:39Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Extract ayo-todos plugin from current todo tool

Extract the current todo tool implementation into the ayo-todos planner plugin. This becomes the default near-term planner.

## Context
- Current implementation: internal/tools/todos/todos.go
- Target: internal/planners/builtin/todos/
- This is the reference implementation for near-term planners

## Current Tool Analysis
The existing todos tool in internal/tools/todos/todos.go provides:
- todos tool with create/update/complete operations
- State stored in session context (not persisted)

## Implementation
1. Create internal/planners/builtin/todos/plugin.go implementing PlannerPlugin
2. Move tool definition logic
3. Implement state persistence in StateDir
4. Register with DefaultRegistry in init()

## Files to Create
- internal/planners/builtin/todos/plugin.go
- internal/planners/builtin/todos/state.go
- internal/planners/builtin/todos/tools.go

## Files to Modify
- None (keep old tool for backwards compat initially)

## Dependencies
- am-tzub (PlannerRegistry)

## Acceptance
- ayo-todos plugin implements PlannerPlugin
- Provides same todos tool functionality
- State persists in sandbox directory
- Registered as 'ayo-todos' in registry


## Notes

**2026-02-18T03:24:09Z**

Split into atomic tickets: am-wil0, am-efcd, am-0011, am-rozh
