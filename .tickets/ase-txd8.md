---
id: ase-txd8
status: closed
deps: []
links: []
created: 2026-02-07T03:22:36Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-7l1g
---
# Standardize trigger ID format and picker

Improve trigger ID handling across commands.

Current: trig_123456789 (random suffix)

Issues:
- Hard to remember/type full IDs
- No auto-complete or picker

Changes:
1. Support short ID prefix matching (like sandbox commands)
2. Add interactive picker when ID not provided
3. Consider shorter IDs or user-provided names

Examples:
  ayo triggers show trig_abc  # prefix match
  ayo triggers rm             # shows picker if multiple
  ayo cron @backup "hourly" --name backup-hourly  # custom name

## Acceptance Criteria

- Short ID prefixes work for all trigger commands
- Picker shown when ID omitted and multiple triggers exist
- Auto-select when only one trigger exists

