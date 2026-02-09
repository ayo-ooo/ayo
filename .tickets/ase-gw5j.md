---
id: ase-gw5j
status: open
deps: []
links: []
created: 2026-02-09T03:03:21Z
type: epic
priority: 2
assignee: Alex Cabrera
parent: ase-zlew
---
# Dynamic Agent Creation

Enable @ayo to create, refine, and manage specialized agents. Agents are created when patterns emerge, refined based on usage, and can be promoted to user-owned.

## Acceptance Criteria

- @ayo can create agents with generated system prompts
- Agent creation tracked in SQLite
- Agents refined based on usage patterns
- Invocation context (ephemeral instructions) works without creating agents
- Users can promote @ayo-created agents
- Clean naming (no dots, no session IDs)

