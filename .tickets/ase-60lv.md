---
id: ase-60lv
status: closed
deps: [ase-trs4]
links: []
created: 2026-02-07T03:25:35Z
type: feature
priority: 3
assignee: Alex Cabrera
parent: ase-6khq
---
# Add natural language schedule parsing

Allow natural language for schedule command (nice-to-have).

Examples:
  ayo trigger schedule @backup "every hour"
  ayo trigger schedule @reports "every day at 9am"
  ayo trigger schedule @cleanup "every monday at 3pm"

Implementation:
- Parse common patterns to cron syntax
- Fallback to literal cron if not recognized
- Consider using existing Go library

Lower priority - cron syntax works fine for power users.

## Acceptance Criteria

- Common patterns work: hourly, daily, weekly
- "every day at Xam/pm" works
- Falls back to cron syntax gracefully
- Clear error for unrecognized input

