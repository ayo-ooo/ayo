---
id: am-lhvw
status: open
deps: [am-gt91, am-khpe]
links: []
created: 2026-02-20T02:51:03Z
type: feature
priority: 2
assignee: Alex Cabrera
tags: [squads, delegation]
---
# Implement agent-to-agent delegation within squad context

Squad lead (@ayo) should be able to delegate work to other agents within the squad. Currently there's no mechanism for one agent to invoke another agent in the same squad sandbox.

## Design

Add a 'delegate' tool that allows the squad lead to invoke another agent defined in SQUAD.md. The delegate tool should: 1) Validate target agent is in squad, 2) Create a sub-session, 3) Run target agent with prompt and squad context, 4) Return result to lead agent.

## Acceptance Criteria

- Squad lead can use delegate tool to invoke @reviewer
- Delegated agent runs in same squad sandbox
- Results are returned to lead agent's context

