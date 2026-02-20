---
id: am-9aco
status: open
deps: []
links: []
created: 2026-02-20T02:49:31Z
type: bug
priority: 1
assignee: Alex Cabrera
tags: [sandbox, shares]
---
# Share system symlinks don't work inside sandbox containers

The share system creates symlinks on the host at ~/.local/share/ayo/sandbox/workspace/{name} pointing to the shared host path. However, when mounted into a container via VirtioFS, symlinks point to host paths (e.g., /tmp/ayo-test-share) that don't exist inside the container. The share appears at /workspace/{name} but resolves to a broken symlink.

## Design

Replace symlink-based shares with direct mounts. When a share is added, dynamically add a mount to the @ayo sandbox configuration. Options: 1) Require sandbox restart for new shares, 2) Implement hot-mount via Apple Container APIs if supported, 3) Mount the parent directories directly instead of using symlinks.

## Acceptance Criteria

- User can add a share with 'ayo share add /path/to/dir --as name'
- The shared directory is immediately accessible at /workspace/name inside the @ayo sandbox
- Files can be read and written through the mount
- Share removal properly cleans up the mount

