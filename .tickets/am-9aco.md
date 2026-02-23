---
id: am-9aco
status: closed
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

## Resolution

Replaced symlink-based shares with direct VirtioFS mounts.

### Changes Made

| File | Changes |
|------|---------|
| internal/sandbox/ayo.go | Load shares via `share.NewService().List()` and add each as a separate VirtioFS mount to `/workspace/{name}` |
| internal/share/share.go | Removed symlink creation in `Add()`, removed symlink deletion in `Remove()` and `RemoveSessionShares()` |
| cmd/ayo/share.go | Updated Long descriptions, added restart hint in output messages |

### How It Works Now

1. **`ayo share add /path`** stores share metadata in shares.json (no symlink)
2. **Sandbox creation** (EnsureAyoSandbox) loads shares and adds each as a direct VirtioFS mount
3. **Inside container** share is available at `/workspace/{name}` via VirtioFS (no symlinks involved)
4. **Restart required** after adding/removing shares (user is informed with message)

### Key Code Locations

| File | Lines | Purpose |
|------|-------|---------|
| internal/sandbox/ayo.go | 103-113 | Share loading and mount creation |
| internal/share/share.go | 198-220 | Add() stores metadata only |
| internal/share/share.go | 253-263 | Remove() updates metadata only |

## Acceptance Criteria

- ✅ User can add a share with 'ayo share add /path/to/dir --as name'
- ✅ After sandbox restart, shared directory is accessible at /workspace/name inside @ayo sandbox
- ✅ Files can be read and written through the mount (VirtioFS supports read/write)
- ✅ Share removal properly cleans up (after restart)
- ✅ Clear messaging when restart is needed
