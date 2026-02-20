---
id: am-pdw1
status: closed
deps: []
links: []
created: 2026-02-20T02:49:39Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [planners, sandbox]
---
# Call InitAyoPlanners during @ayo sandbox initialization

The function InitAyoPlanners exists in internal/sandbox/ayo.go but is never called. This means planners are never initialized for the @ayo sandbox, so planner tools (todos, tickets) are not available to the agent.

## Design

In internal/run/run.go, after ensureAyoSandbox returns successfully, call sandbox.InitAyoPlanners(plannerManager) to initialize the near-term and long-term planners. Store the returned SandboxPlanners for use when building the agent's tool set.

## Acceptance Criteria

- InitAyoPlanners is called when @ayo sandbox is created
- The SandboxPlanners instance is available for tool injection
- Planner state directories exist at /.planner.near and /.planner.long inside sandbox

