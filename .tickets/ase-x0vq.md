---
id: ase-x0vq
status: closed
deps: [ase-95o4, ase-py58]
links: []
created: 2026-02-06T04:09:22Z
type: epic
priority: 1
assignee: Alex Cabrera
parent: ase-95o4
---
# CLI Enhancements for Sandbox

Update all CLI commands to work with the sandbox-first architecture, supporting both human and agentic use.

## Design

## New Commands
- ayo sandbox login [--as @agent]
- ayo sandbox status
- ayo mount/unmount commands
- ayo backup/restore commands
- ayo sync commands
- ayo triggers commands
- ayo messages (IRC log access)
- ayo agents status/wake/sleep

## Enhanced Commands
- ayo sessions continue --latest
- --latest flag on session-related commands
- --mount flag on root command
- --json/--quiet/--yes flags everywhere

## Agentic Considerations
- Meaningful exit codes
- JSON output option
- Stdin/stdout piping
- No interactive prompts with --yes

## Acceptance Criteria

- All new commands implemented
- --latest flag works on session commands
- --json output on all commands
- Exit codes documented
- ayo sandbox login excluded from agent tooling

