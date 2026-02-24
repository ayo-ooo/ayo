---
id: ayo-gtm8
status: closed
deps: []
links: []
created: 2026-02-24T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-gtm1
tags: [gtm, cleanup, dependencies]
---
# Task: Cron Dependency Cleanup

## Summary

The project may have duplicate cron dependencies. PLAN.md indicates migration from `robfig/cron` to `go-co-op/gocron`, but both might still be in go.mod.

## Current State

**Need to verify:**
```bash
grep -E "(robfig/cron|go-co-op/gocron)" go.mod
```

## Analysis Required

### If both exist:

1. Determine which is actively used
2. Search for imports of each
3. If robfig/cron is unused, remove it
4. If migration is incomplete, complete it or document as tech debt

### Import search:
```bash
grep -rn "robfig/cron" ./internal/ ./cmd/
grep -rn "go-co-op/gocron" ./internal/ ./cmd/
```

## Implementation Steps

1. [ ] Check `go.mod` for both cron libraries
2. [ ] Search codebase for imports of each
3. [ ] If robfig/cron unused:
   - Remove from go.mod
   - Run `go mod tidy`
4. [ ] If robfig/cron still used:
   - Decide: complete migration or keep both
   - Document decision
5. [ ] Run tests to verify trigger system still works
6. [ ] Test `ayo trigger` commands manually

## Acceptance Criteria

- [ ] Only one cron library in use (or documented reason for both)
- [ ] `go mod tidy` makes no changes
- [ ] All trigger-related tests pass
- [ ] `ayo trigger add/list/remove` work correctly
