---
id: ase-89e9
status: closed
deps: [ase-0r4v]
links: []
created: 2026-02-10T01:36:49Z
type: task
priority: 3
assignee: Alex Cabrera
parent: ase-8d04
---
# Remove grant loading from sandbox creation (future)

Remove the grant loading code from sandbox creation after the deprecation period.

## Context
Once users have migrated to shares and the mount command has been removed, we can simplify sandbox creation by removing grant loading entirely.

## IMPORTANT: This is a FUTURE ticket
Do NOT implement until:
1. At least one release with deprecation warnings has shipped
2. 'ayo share migrate' has been available for user migration
3. Decision made to proceed with removal

## Files to Modify
- internal/server/sandbox_manager.go
- internal/sandbox/pool.go
- cmd/ayo/mount.go (or delete entirely)

## Changes

### sandbox_manager.go
Remove:
- Import of internal/sandbox/mounts
- Call to mounts.LoadGrants()
- Loop adding grants.ToProviderMounts() to mount list
- Any grant validation code

### pool.go
Remove:
- Any grant-related mount configuration
- Grant loading if present

### mount.go
Either:
- Delete the file entirely
- Or keep as stub that just shows error message

## Before/After

### Before (createPersistentSandbox)
```go
// Load grants
grants, err := mounts.LoadGrants()
if err != nil {
    return err
}

// Add grant mounts
for _, gm := range grants.ToProviderMounts() {
    mounts = append(mounts, providers.Mount{
        Source:      gm.Source,
        Destination: gm.Destination,
        // ...
    })
}
```

### After
```go
// Static mounts only - grants system removed
// Workspace mount handles user-shared paths
```

## Verification
After removal:
1. Sandboxes still start correctly
2. /workspace/ is accessible
3. Shares work as expected
4. No references to mounts.json remain in sandbox code

## Acceptance Criteria

- [ ] Grant loading removed from sandbox_manager.go
- [ ] Grant loading removed from pool.go
- [ ] mount.go deleted or made into error stub
- [ ] mounts package can be deleted
- [ ] Sandboxes start without grants
- [ ] All tests pass
- [ ] Documentation updated

