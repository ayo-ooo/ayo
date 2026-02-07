---
id: ayo-f6ac
status: open
deps: []
links: []
created: 2026-02-06T22:25:16Z
type: bug
priority: 2
assignee: Alex Cabrera
tags: [cli, triggers]
---
# Missing 'ayo triggers' CLI commands

## Issue
AGENT_MANUAL_TEST.md Section 7 expects 'ayo triggers' subcommands:
- `ayo triggers list`
- `ayo triggers add --type cron|watch --agent <agent> --schedule|--path <value> --prompt <prompt>`
- `ayo triggers test <id>`
- `ayo triggers enable/disable <id>`
- `ayo triggers rm <id>`

These commands do not exist.

## Current State
- No 'ayo triggers' command in CLI
- AGENTS.md references trigger_engine.go for 'Cron/watch trigger handling'

## Impact
Cannot create, list, or manage automated triggers via CLI.

