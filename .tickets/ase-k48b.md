---
id: ase-k48b
status: open
deps: []
links: []
created: 2026-02-09T03:03:27Z
type: epic
priority: 2
assignee: Alex Cabrera
parent: ase-zlew
---
# Flow System

Implement flows as saved orchestration patterns. Flows are YAML files that @ayo generates after proving a pattern works. They support shell commands and agent invocations with triggers.

## Acceptance Criteria

- Flow spec supports shell and agent steps
- Flows created by @ayo, editable by users
- Execution stats tracked in SQLite
- Triggers defined in flow files
- Conditional steps supported
- ayo flows CLI for management

