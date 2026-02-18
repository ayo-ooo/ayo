---
id: am-8hkb
status: closed
deps: [am-2u05, am-kjh7, am-rozh, am-eh10]
links: []
created: 2026-02-18T03:14:17Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Inject planner tools into agent tool set

Modify the agent runner to inject planner tools into the agent's available tools.

## Context
- Agent tool setup: internal/run/run.go (toolset building)
- Planners expose tools via Tools() method
- Tools should be injected based on sandbox's planners

## Implementation
When building agent toolset:
1. Get planners for current sandbox via SandboxPlannerManager
2. Call nearTermPlanner.Tools() and longTermPlanner.Tools()
3. Add to agent's tool definitions
4. Handle name collisions (error or namespace)

## Files to Modify
- internal/run/run.go (tool building logic)

## Dependencies
- am-2u05 (SandboxPlannerManager)
- am-tgzo (ayo-todos with Tools())
- am-oqoy (ayo-tickets with Tools())

## Acceptance
- Planner tools appear in agent's available tools
- Tools work within sandbox context
- No tool name collisions (or graceful handling)

