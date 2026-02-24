---
id: ayo-gtm6
status: closed
deps: []
links: []
created: 2026-02-24T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-gtm1
tags: [gtm, modernization, code-quality]
---
# Task: gopls Modernization (172+ Hints)

## Summary

gopls reports 172+ modernization hints across the codebase. While not errors, these represent opportunities to use newer Go idioms and improve code quality for GTM.

## Categories of Hints

### 1. `interface{}` → `any` (efaceany)

Go 1.18 introduced `any` as an alias for `interface{}`. Modern Go code should use `any`.

**Files affected:**
- `cmd/ayo/backup.go:147`
- `cmd/ayo/flows.go:245, 800`
- `cmd/ayo/memory.go:1003, 1091, 1140, 1147`
- Many more throughout `internal/`

### 2. String Processing Modernization (stringsseq)

Go 1.24 introduced `strings.SplitSeq` which is more efficient for range loops.

**Files affected:**
- `cmd/ayo/flows.go:322, 778`
- Various other locations

### 3. `min`/`max` Builtins (minmax)

Go 1.21 introduced `min()` and `max()` builtins to replace manual if statements.

**Files affected:**
- `cmd/ayo/flows.go:334`
- Various other locations

### 4. Other Modernizations

- `slices.Sort` instead of `sort.Slice` where applicable
- `maps.Keys` instead of manual key extraction
- Loop variable modernization (Go 1.22)

## Implementation Plan

### Option 1: Automated Fix

```bash
# gopls can auto-fix many of these
gopls fix -all ./...
```

### Option 2: Manual Batch Fix

Address by category in separate commits:

1. `interface{}` → `any` (sed/find-replace)
2. `min`/`max` usage (manual review)
3. `strings.SplitSeq` (manual review - not always applicable)
4. Other modernizations (case-by-case)

## Implementation Steps

1. [ ] Run `gopls fix -all ./...` and review changes
2. [ ] Manually review any changes that gopls couldn't fix
3. [ ] Run tests after each batch
4. [ ] Verify zero gopls hints remaining

## Acceptance Criteria

- [ ] Zero gopls modernization hints
- [ ] All tests pass
- [ ] Code uses modern Go idioms throughout
