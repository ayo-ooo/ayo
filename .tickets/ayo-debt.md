---
id: ayo-debt
status: open
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-6h19
tags: [cleanup, tech-debt]
---
# Apply gopls modernize suggestions

Apply all gopls modernization hints to clean up the codebase.

## Changes (80+ hints)

### slicescontains (7 locations)
Replace manual loops with `slices.Contains`:

```go
// Before
for _, v := range list {
    if v == target { return true }
}
return false

// After
return slices.Contains(list, target)
```

**Files:**
- `internal/agent/agent.go:409`
- `internal/agent/agent.go:850`
- `internal/memory/zettelkasten/provider.go` (5 locations)

### omitzero (5 locations)
Fix `omitempty` on nested struct fields (has no effect):

**Files:**
- `internal/agent/agent.go:89,97,133,170,171`

Options:
1. Remove `omitempty` tag
2. Use pointer type with `omitempty`
3. Use `omitzero` tag if json v2

### mapsloop (1 location)
Replace loop with `maps.Copy`:

**File:** `internal/flows/yaml_executor.go:471`

### stringscutprefix (1 location)
Use `CutPrefix` instead of `HasPrefix+TrimPrefix`:

**File:** `cmd/ayo/root.go:557`

### stringsseq (1 location)
Use `SplitSeq` for ranging:

**File:** `cmd/ayo/root.go:401`

### fmtappendf (1 location)
Use `fmt.Appendf`:

**File:** `internal/daemon/server.go:834`

## Implementation

1. Add `slices` and `maps` imports where needed
2. Apply each change
3. Run `go test ./...` after each batch
4. Run `golangci-lint run` to verify

## Testing

- All existing tests pass
- No behavior changes
