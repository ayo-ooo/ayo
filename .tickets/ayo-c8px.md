---
id: ayo-c8px
status: closed
deps: []
links: []
created: 2026-02-23T22:15:18Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [sandbox, mount]
---
# Add host home mount at /mnt/{username}

Mount the user's host home directory to /mnt/{username} in read-only mode. This gives agents read access to host files without write permission.

## Implementation

1. Detect current user's home directory
2. Add mount configuration to Apple Container provider
3. Add mount configuration to systemd-nspawn provider
4. Mount as read-only to prevent accidental writes
5. Use /mnt/{username} as mount point (e.g., /mnt/alex)

## Acceptance Criteria

- [ ] Agent can read files from host home directory
- [ ] Agent CANNOT write to host home directory
- [ ] Works on both macOS (Apple Container) and Linux (nspawn)
- [ ] Mount point follows /mnt/{username} convention
- [ ] No sensitive host paths exposed (e.g., .ssh keys stay readable)

## Files to Modify

- `internal/sandbox/apple/container.go`
- `internal/sandbox/nspawn/container.go`

