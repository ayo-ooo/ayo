---
id: ase-76ox
status: closed
deps: [ase-95o4]
links: []
created: 2026-02-06T04:08:58Z
type: epic
priority: 1
assignee: Alex Cabrera
parent: ase-95o4
---
# Mount and Permission System

Implement the mount permission system for controlled host filesystem access from sandbox.

## Design

## Mount Types
1. Persistent grants (mounts.json)
2. Project config (.ayo.json mounts section)
3. Session-scoped (--mount flag)

## CLI Commands
- ayo mount <path> [--readonly]
- ayo mount list
- ayo mount revoke <path>
- ayo --mount <path> 'prompt'

## Resolution Order
1. CLI flags (session-scoped)
2. Project config (.ayo.json)
3. Persistent grants (mounts.json)

## Headless Mode
Flows/triggers declare required mounts upfront. Missing grants fail with clear error message.

## Acceptance Criteria

- ayo mount command manages persistent grants
- Project .ayo.json can declare mounts
- --mount flag for session-scoped access
- Clear errors when mount not granted
- Flows declare mounts in frontmatter

