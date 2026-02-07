---
id: ase-19og
status: open
deps: [ase-trs4, ase-8g9z]
links: []
created: 2026-02-07T03:25:22Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-6khq
---
# Remove `trigger add` command

Remove the generic `add` command now that we have `schedule` and `watch`.

Current:
  ayo triggers add --type cron --agent @backup --schedule "..."
  ayo triggers add --type watch --agent @build --path ./src

After:
  ayo trigger schedule @backup "..."
  ayo trigger watch ./src @build

Changes:
- Remove addTriggerCmd() or make it hidden
- Update help text to show schedule/watch
- Keep for backwards compat if needed (hidden)

## Acceptance Criteria

- `ayo trigger add` is hidden or removed
- Help shows schedule and watch as the way to create triggers
- Old scripts using `trigger add` still work (hidden alias)

