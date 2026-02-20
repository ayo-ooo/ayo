---
id: am-y55g
status: closed
deps: [am-pdw1]
links: []
created: 2026-02-20T02:49:46Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [planners, tools]
---
# Pass PlannerTools to NewFantasyToolSet in buildFantasyAgent

In internal/run/run.go buildFantasyAgent function, NewFantasyToolSet is called without PlannerTools. The ToolSetOptions struct has a PlannerTools field but it's never populated. This prevents planner tools (todos, tickets) from being available to agents.

## Design

After planners are initialized, call GetPlannerTools(nearTerm, longTerm) and pass the result to ToolSetOptions.PlannerTools when creating the tool set.

## Acceptance Criteria

- PlannerTools is populated in ToolSetOptions
- Agent has access to 'todos' tool from near-term planner
- Agent has access to ticket tools from long-term planner
- Running 'ayo "Create a todo list"' results in todos tool usage

