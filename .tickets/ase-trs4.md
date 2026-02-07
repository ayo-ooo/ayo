---
id: ase-trs4
status: open
deps: []
links: []
created: 2026-02-07T03:25:11Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ase-6khq
---
# Add `trigger schedule` subcommand

Add `schedule` subcommand for creating cron-based triggers.

Current:
  ayo triggers add --type cron --agent @backup --schedule "0 0 2 * * *"

Proposed:
  ayo trigger schedule @backup "0 0 2 * * *"
  ayo trigger schedule @backup "0 0 2 * * *" --prompt "Run backup"

Usage: trigger schedule <agent> <schedule> [flags]

Flags:
  --prompt    Prompt to pass to agent
  --name      Custom trigger name/ID

The schedule arg accepts cron syntax (6 fields with seconds).

## Acceptance Criteria

- `ayo trigger schedule @agent "0 * * * * *"` creates trigger
- Shows created trigger ID on success
- --prompt flag works
- Error if agent or schedule missing

