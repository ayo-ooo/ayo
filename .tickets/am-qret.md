---
id: am-qret
status: open
deps: [am-2u05, am-rozh, am-eh10]
links: []
created: 2026-02-18T03:14:24Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Inject planner instructions into system prompt

Modify agent system prompt building to inject planner instructions.

## Context
- System prompt building: internal/agent/agent.go (CombinedSystem)
- Planners expose instructions via Instructions() method
- Instructions teach agent how to use planner tools

## Implementation
When building combined system prompt:
1. Get planners for current sandbox
2. Call nearTermPlanner.Instructions() and longTermPlanner.Instructions()
3. Inject into system prompt (similar to skill injection)

## Files to Modify
- internal/agent/agent.go or internal/run/run.go (system prompt building)

## Dependencies
- am-2u05 (SandboxPlannerManager)
- am-tgzo (ayo-todos with Instructions())
- am-oqoy (ayo-tickets with Instructions())

## Acceptance
- Planner instructions appear in system prompt
- Instructions scoped to sandbox context
- Clear separation between near/long-term instructions

