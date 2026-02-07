---
id: ayo-2s3e
status: open
deps: []
links: []
created: 2026-02-06T22:23:06Z
type: bug
priority: 2
assignee: Alex Cabrera
tags: [cli, mount]
---
# Missing 'ayo mount' CLI commands

## Issue
AGENT_MANUAL_TEST.md Section 5 expects 'ayo mount' subcommands:
- `ayo mount list`
- `ayo mount add <path> [--readonly] --reason <reason>`
- `ayo mount rm <path>`

These commands do not exist.

## Current State
- No 'ayo mount' command in CLI
- No mounts.json file exists
- AGENTS.md references ~/.local/share/ayo/mounts.json as 'Persistent mount permissions'

## Expected
CLI commands to manage mounts for sandbox access to host directories.

## Workaround
Current working directory is automatically mounted via virtiofs (seen in sandbox mount output).

## Impact
Cannot manually add/remove/list mount permissions via CLI.

