---
id: ase-nvjz
status: closed
deps: [ase-uwnw]
links: []
created: 2026-02-10T01:32:49Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-8d04
---
# Add workspace mount to sandbox manager

Mount the workspace directory into sandboxes at /workspace/ so that symlinks are visible inside containers.

## Context
The sandbox manager creates persistent sandboxes with various mounts. We need to add the workspace directory mount so shared host paths are accessible.

## File to Modify
- internal/server/sandbox_manager.go

## Location
Find the createPersistentSandbox() function. Look for where mounts are configured, likely near lines that mount:
- sync.HomesDir() -> /home
- sync.SharedDir() -> /shared

## Implementation
Add this mount alongside the existing ones:

```go
{
    Source:      sync.WorkspaceDir(),
    Destination: "/workspace",
    Mode:        providers.MountModeBind,
    ReadOnly:    false,
},
```

## Verification
After making this change:
1. The sandbox should have /workspace/ directory accessible
2. Symlinks in WorkspaceDir() on host should be followed
3. Files should be readable/writable through the mount

## Import Check
Ensure internal/sync is imported (it should already be for HomesDir/SharedDir).

## Search Hints
- Search for "sync.HomesDir" to find where mounts are defined
- Search for "Destination.*home" or "Destination.*shared"
- The mount structure likely uses providers.Mount type

## Acceptance Criteria

- [ ] Workspace mount added to createPersistentSandbox()
- [ ] Mount source is sync.WorkspaceDir()
- [ ] Mount destination is /workspace
- [ ] Mount is read-write (ReadOnly: false)
- [ ] Build succeeds
- [ ] Sandbox has /workspace/ directory after creation

