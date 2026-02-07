---
id: ase-t962
status: closed
deps: [ase-4sre, ase-fegj]
links: []
created: 2026-02-07T03:22:20Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-7l1g
---
# Simplify `ayo triggers` to management-only

Refactor triggers command to focus on trigger management, not creation.

After cron/watch commands exist, triggers becomes:
- `ayo triggers` or `ayo triggers list` - list all triggers
- `ayo triggers show <id>` - show details
- `ayo triggers rm <id>` - remove trigger
- `ayo triggers test <id>` - fire manually
- `ayo triggers enable/disable <id>` - toggle

Remove or hide:
- `ayo triggers add` - deprecated, use `ayo cron` or `ayo watch`

Changes:
- Update help text to point to cron/watch
- Keep add command but mark as hidden
- Update examples throughout

## Acceptance Criteria

- `ayo triggers` shows list by default
- `ayo triggers add` still works but is hidden from help
- Help text mentions `ayo cron` and `ayo watch`

