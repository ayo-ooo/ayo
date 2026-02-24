---
id: ayo-xhox
status: closed
deps: [ayo-enaj]
links: []
created: 2026-02-23T22:16:36Z
type: task
priority: 3
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [deps, cleanup]
---
# Update go.mod dependencies

Remove unused dependencies after code removal (Phase 1). Update remaining dependencies to latest versions. Clean up go.mod.

## Context

After Phase 1 code removals, go.mod will have unused dependencies. This ticket cleans up and updates dependencies.

## Steps

### 1. Run go mod tidy

```bash
cd /Users/acabrera/Code/ayo-ooo/ayo
go mod tidy
```

This removes unused dependencies automatically.

### 2. Check for Outdated Dependencies

```bash
go list -m -u all
```

### 3. Update Dependencies

Update key dependencies:
```bash
# Update all direct dependencies
go get -u ./...

# Or update specific ones
go get -u github.com/spf13/cobra
go get -u github.com/go-co-op/gocron/v2
```

### 4. Verify Build

```bash
go build ./cmd/ayo/...
go test ./... -count=1
```

## Dependencies to Remove (Expected)

After Phase 1 removals, these may be unused:
- Libraries only used by removed internal/server/
- Libraries only used by removed flow system
- Libraries only used by removed plugins

## Dependencies to Add (Expected)

Phase 4 adds:
- `github.com/go-co-op/gocron/v2` - Scheduler
- `github.com/santhosh-tekuri/jsonschema/v5` - JSON Schema validation

Phase 2 may add:
- `github.com/bmatcuk/doublestar/v4` - Glob patterns

## Acceptance Criteria

- [ ] `go mod tidy` completes without errors
- [ ] No unused dependencies in go.mod
- [ ] All direct dependencies updated to latest stable
- [ ] Build succeeds after updates
- [ ] Tests pass after updates
- [ ] go.sum updated and committed

## Testing

- `go build ./...` succeeds
- `go test ./... -count=1` passes
- `go mod verify` passes
