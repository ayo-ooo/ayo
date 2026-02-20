---
id: am-gt91
status: in_progress
deps: []
links: []
created: 2026-02-20T02:49:58Z
type: feature
priority: 1
assignee: Alex Cabrera
tags: [squads, dispatch]
---
# Implement actual agent invocation in Squad.Dispatch

Squad.Dispatch in internal/squads/dispatch.go (line 144) has a TODO comment indicating that actual agent invocation is not implemented. Currently it only returns a routing message but doesn't actually run the target agent with the prompt.

## Design

Use the AgentInvoker (am-2wpf, already closed) to invoke the target agent within the squad sandbox context. The dispatch should: 1) Determine target agent via GetTargetAgent, 2) Load the agent configuration, 3) Inject SQUAD.md into system prompt, 4) Run the agent in the squad sandbox, 5) Return the agent's response.

## Acceptance Criteria

- 'ayo "#dev-team" "What is your mission?"' invokes @ayo in squad context
- Agent receives SQUAD.md injected into system prompt
- Agent runs inside squad sandbox with access to /workspace
- Response is returned to user

