---
id: ase-wvrr
status: closed
deps: [ase-ji7h]
links: []
created: 2026-02-06T04:14:46Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-x0vq
---
# Add triggers CLI commands

Implement 'ayo triggers' CLI commands for managing trigger registrations.

## Design

## Commands
ayo triggers list              # List all triggers
ayo triggers show <id>         # Show trigger details
ayo triggers add               # Add trigger (interactive or flags)
ayo triggers remove <id>       # Remove trigger
ayo triggers history           # Show recent trigger fires
ayo triggers test <id>         # Manually fire a trigger

## triggers list
Output:
  ID       Type   Agent     Pattern         Status
  abc123   cron   @ayo      '0 9 * * *'     active
  def456   watch  @crush    ./src/**/*.go   active
  ghi789   irc    @ayo      @ayo            paused

## triggers add
Flags:
  --type cron|watch|webhook|irc
  --agent @ayo
  --pattern '...'
  --prompt 'Trigger context: {context}'

Or interactive mode without flags.

## triggers history
Output:
  Time           Trigger  Agent   Status
  10:30 today    abc123   @ayo    completed
  09:00 today    abc123   @ayo    completed
  08:45 today    def456   @crush  failed: agent busy

## Implementation
cmd/ayo/triggers.go:
- All trigger commands
- Daemon client calls for runtime operations
- File manipulation for config-based triggers

## Acceptance Criteria

- All trigger commands work
- Triggers added/removed correctly
- History shows recent fires
- Test command fires manually

