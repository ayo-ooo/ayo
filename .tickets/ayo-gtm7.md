---
id: ayo-gtm7
status: open
deps: []
links: []
created: 2026-02-24T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-gtm1
tags: [gtm, cleanup, dead-code]
---
# Task: Remove Dead Code

## Summary

Several files contain dead or deprecated code that should be removed for GTM cleanliness.

## Known Dead Code

### 1. `cmd/ayo/chat.go`

**Status:** Likely dead - needs verification

**Indicators:**
- Not referenced in main command tree
- Possibly superseded by `session.go` or `run.go`
- May contain deprecated chat implementation

**Verification:**
```bash
# Check if chat.go defines any commands that are wired
grep -n "chat" cmd/ayo/root.go
grep -rn "chatCmd" cmd/ayo/
```

### 2. Unused Exports

Scan for exported functions/types with no external references:
```bash
# Find potentially unused exports
go-unused ./...  # if available
```

### 3. Commented Code Blocks

Search for large commented-out code blocks that should be deleted:
```bash
grep -rn "// TODO: remove" ./internal/
grep -rn "DEPRECATED" ./internal/
```

### 4. Unused Imports

gopls should catch these, but verify clean:
```bash
go mod tidy
goimports -w ./...
```

## Implementation Steps

1. [ ] Verify `cmd/ayo/chat.go` is unused and remove
2. [ ] Run `go mod tidy` to clean dependencies
3. [ ] Search for and remove large commented-out blocks
4. [ ] Remove any TODO:remove markers and associated code
5. [ ] Run full test suite
6. [ ] Verify build succeeds

## Acceptance Criteria

- [ ] No dead command files
- [ ] No large commented-out code blocks
- [ ] `go mod tidy` makes no changes
- [ ] All tests pass
