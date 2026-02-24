---
id: ayo-2e0t
status: open
deps: []
links: []
created: 2026-02-23T22:16:25Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [memory, docs]
---
# Update AGENTS.md memory file

Update the AGENTS.md memory file to reflect the simplified architecture after all GTM changes.

## Context

AGENTS.md is the primary memory file for AI agents working on the codebase. It needs to accurately reflect the post-GTM architecture.

## Sections to Update

### Project Overview

Update to reflect:
- ayod in-sandbox daemon
- Simplified architecture (no flows, reduced plugins)
- Clear sandbox/host split

### Documentation Map

Update to reflect:
- Removed docs (flows-spec.md, etc.)
- New docs (triggers.md, patterns/)
- Consolidated docs

### Key Directories

Update to include:
- internal/ayod/ - In-sandbox daemon
- internal/daemon/triggers/ - Trigger engine
- Removed directories no longer present

### Key Concepts

Update to include:
- ayod daemon concept
- Real Unix users in sandboxes
- file_request flow
- Trigger types

### Common Commands

Verify all commands still work:
- Build command
- Test command
- Lint command

### Code Conventions

Update any conventions that changed:
- New patterns for sandbox communication
- ayod RPC patterns

### Debugging

Update with new debug scenarios:
- ayod troubleshooting
- Trigger debugging

## New Content to Add

### ayod Section

```markdown
## ayod (In-Sandbox Daemon)

Lightweight daemon running as PID 1 inside sandboxes:
- User management (creates agent users)
- Command execution (runs as specified user)
- File request proxy (forwards to host)
- Output sync (/output → host)

Communication: JSON-RPC over /run/ayod.sock
```

### Trigger Engine Section

```markdown
## Trigger Engine

Scheduler for ambient agents:
- Uses gocron v2
- Supports cron, interval, one-time, file-watch
- Jobs persisted in SQLite
- Hot-reload from ~/.config/ayo/triggers/*.yaml
```

## Files to Modify

1. **`/Users/acabrera/Code/ayo-ooo/ayo/AGENTS.md`** - Main update

## Acceptance Criteria

- [ ] All sections reflect current architecture
- [ ] No references to removed features
- [ ] New features documented
- [ ] Directory paths are accurate
- [ ] Commands work when copy-pasted
- [ ] Debugging tips are current

## Testing

- Read through entire file for accuracy
- Verify directory references exist
- Test documented commands
