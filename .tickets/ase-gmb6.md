---
id: ase-gmb6
status: closed
deps: [ase-4sre]
links: []
created: 2026-02-07T03:22:28Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ase-7l1g
---
# Add natural language cron parsing

Allow natural language input for cron schedules.

Examples:
  ayo cron @backup "every hour"
  ayo cron @reports "every day at 9am"
  ayo cron @cleanup "every monday at 3pm"
  ayo cron @sync "every 5 minutes"

Implementation options:
1. Use a Go library like github.com/lnquy/cron for parsing
2. Build simple parser for common patterns
3. Use LLM to convert (probably overkill)

Fallback: if not recognized, treat as literal cron expression

Common patterns to support:
- every N minutes/hours
- every day at HH:MM
- every weekday at HH:MM
- every monday/tuesday/etc at HH:MM
- hourly, daily, weekly, monthly

## Acceptance Criteria

- `ayo cron @agent "every hour"` works
- `ayo cron @agent "every day at 2am"` works
- Invalid natural language shows helpful error
- Raw cron syntax still works as fallback

