---
id: ase-c1y8
status: closed
deps: [ase-uwnw]
links: []
created: 2026-02-10T01:32:58Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-8d04
---
# Add workspace mount to sandbox pool

Ensure pool-managed sandboxes also get the workspace mount for share visibility.

## Context
The sandbox pool pre-creates sandboxes for faster agent startup. These pooled sandboxes need the same workspace mount as directly-created sandboxes.

## File to Modify
- internal/sandbox/pool.go

## Location
Find the createSandbox() function or wherever mounts are configured for pooled sandboxes. Look for:
- References to providers.Mount
- PoolConfig with Mounts field
- Similar mounts for homes/shared directories

## Implementation
Add the workspace mount in the same manner as sandbox_manager.go:

```go
{
    Source:      sync.WorkspaceDir(),
    Destination: "/workspace",
    Mode:        providers.MountModeBind,
    ReadOnly:    false,
},
```

## Import Check
May need to import:
```go
import (
    "github.com/alexcabrera/ayo/internal/sync"
)
```

## Search Hints
- Search for "Mounts" in pool.go
- Search for "providers.Mount"
- Look at PoolConfig struct for mount configuration

## Acceptance Criteria

- [ ] Workspace mount added to pool sandbox creation
- [ ] Mount source is sync.WorkspaceDir()
- [ ] Mount destination is /workspace
- [ ] Mount is read-write
- [ ] Build succeeds
- [ ] Pooled sandboxes have /workspace/ accessible

