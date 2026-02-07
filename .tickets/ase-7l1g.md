---
id: ase-7l1g
status: closed
deps: []
links: []
created: 2026-02-07T03:22:00Z
type: epic
priority: 2
assignee: Alex Cabrera
---
# Triggers CLI Redesign Epic

Redesign the triggers CLI for better UX. Current issues:
- `ayo triggers add --type cron --agent @backup --schedule "0 0 2 * * *"` is verbose
- Mixing cron and watch options in one command is confusing
- No natural language for cron schedules
- No shorthand for common patterns

Target UX:
- `ayo cron @backup "every day at 2am"` - natural language cron
- `ayo watch ./src --agent @build --patterns "*.go"` - simpler watch
- Split into separate `cron` and `watch` top-level commands
- Keep `triggers list/show/rm/test` for management

