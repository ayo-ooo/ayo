---
id: ase-4sre
status: closed
deps: []
links: []
created: 2026-02-07T03:22:08Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ase-7l1g
---
# Add top-level `ayo cron` command

Create a new top-level `cron` command for scheduling agent wakeups.

Current:
  ayo triggers add --type cron --agent @backup --schedule "0 0 2 * * *"

Proposed:
  ayo cron @backup "0 0 2 * * *"
  ayo cron @backup "every day at 2am"    # with natural language (future)

Implementation:
- New cmd/ayo/cron.go file
- Simple positional args: <agent> <schedule>
- Optional --prompt flag for custom prompt
- Optional --id flag to set custom ID
- Calls same daemon.TriggerRegister under the hood
- `triggers add --type cron` becomes hidden alias for backwards compat

## Acceptance Criteria

- `ayo cron @agent "0 * * * * *"` creates a cron trigger
- `ayo cron --help` shows clear usage
- Output shows created trigger ID
- Works with existing daemon trigger system

