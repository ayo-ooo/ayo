---
id: ase-oeok
status: closed
deps: [ase-89e9]
links: []
created: 2026-02-10T01:37:00Z
type: task
priority: 3
assignee: Alex Cabrera
parent: ase-8d04
---
# Remove mount command entirely (future)

Final cleanup: remove the deprecated mount command entirely.

## IMPORTANT: This is a FUTURE ticket
Do NOT implement until:
1. SHARE-032 is complete (grant loading removed)
2. At least 2 releases since deprecation started
3. Announcement made about removal

## Files to Delete/Modify
- cmd/ayo/mount.go (DELETE)
- cmd/ayo/root.go (remove AddCommand for mount)
- internal/sandbox/mounts/ (DELETE directory)
- internal/sandbox/mounts/grants.go (DELETE)
- internal/sandbox/mounts/mounts.go (DELETE if exists)

## Changes to root.go
Remove:
```go
cmd.AddCommand(newMountCmd())
```

## Documentation Updates
- AGENTS.md: Remove mount command section
- README.md: Remove any mount references
- Any other docs mentioning 'ayo mount'

## Verification
1. Build succeeds
2. 'ayo mount' shows 'unknown command' error
3. 'ayo share' works correctly
4. All tests pass
5. No orphaned references to mount/grants

## Acceptance Criteria

- [ ] cmd/ayo/mount.go deleted
- [ ] mount command removed from root.go
- [ ] internal/sandbox/mounts/ deleted
- [ ] Build succeeds
- [ ] 'ayo mount' shows unknown command
- [ ] Documentation updated
- [ ] No orphaned code references

