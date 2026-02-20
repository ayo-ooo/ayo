---
id: am-khpe
status: open
deps: [am-gt91]
links: []
created: 2026-02-20T02:50:14Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [squads, constitution]
---
# Inject SQUAD.md constitution into agent system prompt during squad dispatch

When an agent is invoked within a squad context, the SQUAD.md file should be injected into the agent's system prompt so the agent understands its role, the squad's mission, and coordination rules.

## Design

In the squad dispatch invocation path, after loading the target agent, call squads.InjectConstitution() to add the SQUAD.md content to the system prompt. The constitution should appear in a <squad_context> section.

## Acceptance Criteria

- Agent invoked via '#squad-name' has SQUAD.md in system prompt
- Agent can answer questions about squad mission and roles
- SQUAD.md frontmatter is parsed for planner configuration

