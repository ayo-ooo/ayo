---
id: ayo-5kns
status: open
deps: [ayo-c5mt]
links: []
created: 2026-02-23T22:15:34Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-whmn
tags: [approval, ux]
---
# Add session-scoped approval caching

When user selects "Always for session" in the file approval UI, cache that decision for the remainder of the session. Support patterns to reduce approval fatigue for repetitive operations.

## Context

The file_request approval UI (ayo-c5mt) prompts for every modification. When agents are editing many files (e.g., refactoring), users may want to approve once and let the agent continue. This ticket adds session-scoped caching so approval decisions persist within a session.

This is highest priority in the approval chain (ayo-evik):
1. **Session cache** ← This ticket
2. `--no-jodas` CLI flag
3. Agent-level `permissions.auto_approve`
4. Global `permissions.no_jodas`
5. Prompt user

## Approval Options in UI

Expand the approval UI to offer:
1. **Allow** - Approve this one file
2. **Allow pattern** - Approve files matching a pattern (e.g., `*.md` in current directory)
3. **Always for session** - Auto-approve all future requests this session
4. **Deny** - Reject this request
5. **Deny and abort** - Reject and stop agent execution

## Data Structure

```go
// internal/tools/approval_cache.go
type ApprovalCache struct {
    mu        sync.RWMutex
    patterns  []ApprovalPattern  // Pattern-based approvals
    allFiles  bool               // "Always for session" was selected
}

type ApprovalPattern struct {
    Pattern   string    // glob pattern like "*.md" or "src/**/*.go"
    Directory string    // scoped to this directory
    CreatedAt time.Time
}

func (c *ApprovalCache) IsApproved(path string) bool
func (c *ApprovalCache) AddPattern(pattern, dir string)
func (c *ApprovalCache) ApproveAll()
```

## Files to Modify

1. **`internal/tools/approval_cache.go`** (new) - Implement cache
2. **`internal/tools/file_request.go`** - Check cache before prompting
3. **`internal/ui/approval.go`** - Add new options to UI

## Session Scope

Cache lives in memory, tied to the daemon process or agent session:
- Cleared when agent exits
- Cleared when daemon restarts
- NOT persisted to disk (security)

## Pattern Matching

Use `doublestar` library for glob patterns:
- `*.md` - All markdown files in current directory
- `**/*.go` - All Go files recursively
- `src/` - All files under src/

## Acceptance Criteria

- [ ] "Always for session" option appears in approval UI
- [ ] Selecting it bypasses all future prompts
- [ ] Pattern-based approval works with glob patterns
- [ ] Cache is cleared on session end
- [ ] Cache is NOT persisted to disk
- [ ] Audit logging still records all auto-approved operations

## Testing

- Test cache stores approval decisions
- Test pattern matching works correctly
- Test cache clears on session end
- Test audit logging for auto-approved requests
- Test UI displays new options correctly
