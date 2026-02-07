---
id: ase-ltet
status: closed
deps: [ase-hjhk]
links: []
created: 2026-02-06T04:12:56Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-76ox
---
# Add --mount flag to root command

Add session-scoped mount flag to the root ayo command for temporary filesystem access.

## Design

## Flag
ayo --mount <path> 'prompt'
ayo -m <path> -m <path2> 'prompt'

## Behavior
- Grant temporary access for this session only
- Not persisted to mounts.json
- Combined with persistent and project mounts

## Implementation
1. Add --mount/-m flag to root command (repeatable)
2. Pass mount list to Runner
3. Runner adds to sandbox mounts for this session
4. Mounts cleaned up when session ends

## Mount Resolution Order
1. CLI --mount flags (session-scoped)
2. Project .ayo.json mounts
3. Persistent mounts.json

Higher priority wins for conflicts.

## Validation
- Path must exist (or be creatable)
- Path must be accessible by current user
- Warn if already mounted via other source

## Interactive Mode
In interactive mode, mounts established at start apply to whole session.
Future: could add /mount command in REPL.

## Acceptance Criteria

- --mount flag works on root command
- Multiple -m flags supported
- Session-scoped (not persisted)
- Combined with other mount sources

