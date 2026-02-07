---
id: ase-uals
status: closed
deps: [ase-hjhk]
links: []
created: 2026-02-06T04:12:47Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-76ox
---
# Add mount commands to CLI

Implement 'ayo mount' CLI commands for managing persistent filesystem access grants.

## Design

## Commands
ayo mount <path>              # Grant readwrite access
ayo mount <path> --readonly   # Grant readonly access
ayo mount list                # List all grants
ayo mount revoke <path>       # Remove grant
ayo mount revoke --all        # Remove all grants

## Implementation
cmd/ayo/mount.go:
- newMountCmd() - parent command
- newMountGrantCmd() - default action, grant access
- newMountListCmd() - list grants
- newMountRevokeCmd() - revoke access

## Path Handling
- Resolve relative paths to absolute
- Validate path exists
- Normalize path separators

## Output
mount list:
  /Users/alex/Code/project  readwrite  2025-02-05
  /Users/alex/Documents     readonly   2025-02-01

mount (grant):
  Granted readwrite access to /Users/alex/Code/project

## Flags
--json: Output in JSON format
--quiet: Minimal output

## Error Cases
- Path doesn't exist: warn but allow (may not exist yet)
- Already granted: update mode if different, or no-op
- Not granted (revoke): warn

## Acceptance Criteria

- All mount commands work
- JSON output available
- Path resolution correct
- Integration with mounts service

