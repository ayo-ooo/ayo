---
id: ayo-m1zl
status: closed
deps: []
links: []
created: 2026-02-23T22:16:25Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [cli, help]
---
# Polish CLI help text

Review and improve help text for all ayo commands. Ensure consistency, add examples, and remove mentions of removed features.

## Context

Help text is often the first documentation users see. This ticket ensures all CLI help is clear, consistent, and helpful.

## Commands to Review

### Root Command

```
ayo - AI agents that live on your machine

Usage:
  ayo [command] [flags]
  ayo [prompt]           # Chat with default agent

Common Commands:
  (none)        Chat with @ayo in current directory
  @agent        Chat with specific agent
  #squad        Send to squad

Management Commands:
  agent         Manage agents (create, list, show, delete)
  squad         Manage squads (create, list, shell, delete)
  trigger       Manage triggers (create, list, show, enable, disable)
  daemon        Control background service

Other Commands:
  doctor        Check system health
  share         Manage host directory shares
  dispatch      Send message to agent/squad (non-interactive)

Flags:
  -y, --no-jodas   Auto-approve file modifications
  -q, --quiet      Suppress non-essential output
      --json       Output in JSON format
  -h, --help       Show help

Examples:
  ayo "explain this code"              # Chat with @ayo
  ayo @reviewer "review my changes"    # Chat with @reviewer
  ayo dispatch #dev-team "build auth"  # Send to squad
  ayo trigger create daily-report      # Create trigger
```

### Subcommands

For each subcommand, ensure:
1. One-line description is clear
2. Usage pattern is correct
3. Examples section included
4. Flags documented
5. No references to removed features

## Consistency Rules

- Use imperative mood ("Create", not "Creates")
- Keep descriptions under 60 characters
- Examples use realistic values
- Flags in consistent order (common first)
- Related commands grouped together

## Files to Modify

1. `cmd/ayo/root.go` - Root help
2. `cmd/ayo/agent.go` - Agent commands
3. `cmd/ayo/squad.go` - Squad commands
4. `cmd/ayo/trigger.go` - Trigger commands
5. `cmd/ayo/daemon.go` - Daemon commands
6. `cmd/ayo/share.go` - Share commands
7. `cmd/ayo/doctor.go` - Doctor command
8. `cmd/ayo/dispatch.go` - Dispatch command

## Acceptance Criteria

- [ ] All commands have clear one-line descriptions
- [ ] All commands have usage examples
- [ ] All flags are documented
- [ ] No references to removed features
- [ ] Consistent formatting throughout
- [ ] `ayo help` is useful for new users
- [ ] `ayo help <command>` provides enough detail

## Testing

- Run `ayo help` and verify output
- Run `ayo help <command>` for each command
- Verify examples work when copy-pasted
- Check no broken references
