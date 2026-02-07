---
id: ase-3j4f
status: open
deps: []
links: []
created: 2026-02-07T03:25:04Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-6khq
---
# Rename `triggers` to `trigger` (singular)

Rename the command from `triggers` to `trigger` for consistency.

Most CLI tools use singular: git commit, docker container, kubectl pod.

Changes:
- Rename newTriggersCmd() -> newTriggerCmd()
- Change Use: "triggers" -> "trigger"
- Keep "triggers" as hidden alias for backwards compat
- Update all help text and examples

## Acceptance Criteria

- `ayo trigger` works
- `ayo triggers` still works (hidden alias)
- Help shows `trigger` as primary

