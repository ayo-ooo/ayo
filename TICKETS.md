# Ayo Share System: Implementation Tickets

This document contains all tickets for implementing the share system as described in PLAN.md.

---

## Epic: Share System Implementation

Replace the mount/grants system with a simpler `ayo share` command that uses symlinks for instant file sharing without sandbox restarts.

---

## Phase 1: Core Infrastructure

### SHARE-001: Create share package with core types and service

**Priority**: P0 (Blocker)  
**Estimate**: 2 hours  
**Dependencies**: None

**Description**:
Create the foundational `internal/share/` package with types and service for managing shares.

**Acceptance Criteria**:
- [ ] Create `internal/share/share.go` with:
  - `Share` struct: `Name`, `Path`, `Session`, `SharedAt`
  - `SharesFile` struct: `Version`, `Shares`
  - `ShareService` with mutex for thread safety
- [ ] Implement `NewShareService()` constructor
- [ ] Implement `Load()` - reads from shares.json, handles missing file
- [ ] Implement `Save()` - writes to shares.json with proper permissions
- [ ] Implement `Add(path, name, session)` - adds new share
- [ ] Implement `Remove(nameOrPath)` - removes share by name or original path
- [ ] Implement `List()` - returns all shares
- [ ] Implement `Get(name)` - returns single share by name
- [ ] Implement `GetByPath(path)` - returns share by original path
- [ ] Use `paths.DataDir()` for shares.json location (respects dev mode)
- [ ] Add comprehensive unit tests

**Files**:
- `internal/share/share.go` (new)
- `internal/share/share_test.go` (new)

---

### SHARE-002: Add WorkspaceDir to sync package

**Priority**: P0 (Blocker)  
**Estimate**: 30 minutes  
**Dependencies**: None

**Description**:
Add `WorkspaceDir()` function to the sync package and ensure the directory is created during initialization.

**Acceptance Criteria**:
- [ ] Add `WorkspaceDir()` function returning `{SandboxDir}/workspace`
- [ ] Update `Init()` to create workspace directory alongside homes/shared
- [ ] Add .gitkeep to workspace directory
- [ ] Workspace directory has 0755 permissions
- [ ] Works correctly in both dev mode and production

**Files**:
- `internal/sync/git.go`

---

### SHARE-003: Add workspace mount to sandbox manager

**Priority**: P0 (Blocker)  
**Estimate**: 30 minutes  
**Dependencies**: SHARE-002

**Description**:
Mount the workspace directory into sandboxes at `/workspace/`.

**Acceptance Criteria**:
- [ ] Add mount in `createPersistentSandbox()`:
  ```go
  {
      Source:      sync.WorkspaceDir(),
      Destination: "/workspace",
      Mode:        providers.MountModeBind,
      ReadOnly:    false,
  }
  ```
- [ ] Verify mount appears in sandbox after creation
- [ ] Verify `/workspace/` is writable inside container

**Files**:
- `internal/server/sandbox_manager.go`

---

### SHARE-004: Add workspace mount to sandbox pool

**Priority**: P0 (Blocker)  
**Estimate**: 30 minutes  
**Dependencies**: SHARE-002

**Description**:
Ensure pool-managed sandboxes also get the workspace mount.

**Acceptance Criteria**:
- [ ] Add workspace mount in `createSandbox()` function
- [ ] Mount uses same path as sandbox manager
- [ ] Pooled sandboxes have `/workspace/` accessible

**Files**:
- `internal/sandbox/pool.go`

---

### SHARE-005: Implement symlink operations in share service

**Priority**: P0 (Blocker)  
**Estimate**: 1 hour  
**Dependencies**: SHARE-001, SHARE-002

**Description**:
Add symlink creation and removal to the share service.

**Acceptance Criteria**:
- [ ] `Add()` creates symlink: `WorkspaceDir()/name → absPath`
- [ ] `Add()` validates source path exists (error if not)
- [ ] `Add()` validates name doesn't already exist (error with suggestion)
- [ ] `Add()` auto-generates name from path basename if not provided
- [ ] `Remove()` deletes symlink from workspace directory
- [ ] `Remove()` handles case where symlink doesn't exist (idempotent)
- [ ] All symlink operations are atomic where possible
- [ ] Proper error messages for permission issues

**Files**:
- `internal/share/share.go`
- `internal/share/share_test.go`

---

### SHARE-006: Handle name conflicts in share service

**Priority**: P1 (High)  
**Estimate**: 45 minutes  
**Dependencies**: SHARE-005

**Description**:
Implement proper handling of name conflicts when adding shares.

**Acceptance Criteria**:
- [ ] If name exists, return error with message: `"share 'foo' already exists, use --as to specify a different name"`
- [ ] Implement `GenerateUniqueName(baseName)` for auto-naming if needed
- [ ] Validate name doesn't contain path separators or special chars
- [ ] Validate name isn't empty after normalization
- [ ] Add tests for conflict scenarios

**Files**:
- `internal/share/share.go`
- `internal/share/share_test.go`

---

## Phase 2: CLI Commands

### SHARE-010: Create share command structure

**Priority**: P0 (Blocker)  
**Estimate**: 30 minutes  
**Dependencies**: SHARE-001

**Description**:
Create the base `ayo share` command with subcommand structure.

**Acceptance Criteria**:
- [ ] Create `cmd/ayo/share.go`
- [ ] Add `newShareCmd()` returning cobra.Command
- [ ] Add to root command in `root.go`
- [ ] Command has helpful description and examples
- [ ] `ayo share --help` shows all subcommands

**Files**:
- `cmd/ayo/share.go` (new)
- `cmd/ayo/root.go`

---

### SHARE-011: Implement `ayo share add` command

**Priority**: P0 (Blocker)  
**Estimate**: 1 hour  
**Dependencies**: SHARE-005, SHARE-010

**Description**:
Implement the command to add a new share.

**Acceptance Criteria**:
- [ ] `ayo share <path>` adds share with auto-generated name
- [ ] `ayo share <path> --as <name>` adds share with custom name
- [ ] `ayo share <path> --session` marks share as session-only
- [ ] Supports relative paths, absolute paths, and `~/` expansion
- [ ] Shows success message: `✓ Shared /path/to/project as 'project' → /workspace/project`
- [ ] Shows error if path doesn't exist
- [ ] Shows error if name conflict
- [ ] Supports `--json` output flag
- [ ] No sandbox restart message (that's the point!)

**Example output**:
```
$ ayo share ~/Code/myproject
✓ Shared /Users/alex/Code/myproject → /workspace/myproject
```

**Files**:
- `cmd/ayo/share.go`

---

### SHARE-012: Implement `ayo share list` command

**Priority**: P0 (Blocker)  
**Estimate**: 45 minutes  
**Dependencies**: SHARE-010

**Description**:
Implement the command to list all shares.

**Acceptance Criteria**:
- [ ] `ayo share list` shows table of shares
- [ ] `ayo share ls` alias works
- [ ] Columns: NAME, PATH, TYPE (permanent/session), SHARED
- [ ] Shows helpful message when no shares exist
- [ ] `--json` outputs JSON array
- [ ] Session shares indicated with different styling/icon

**Example output**:
```
  Shares
  ────────────────────────────────────────

  ● project     /Users/alex/Code/project     (2 hours ago)
  ○ temp-data   /tmp/data                    (session, 5 min ago)

  Access at /workspace/{name} inside sandbox
```

**Files**:
- `cmd/ayo/share.go`

---

### SHARE-013: Implement `ayo share rm` command

**Priority**: P0 (Blocker)  
**Estimate**: 45 minutes  
**Dependencies**: SHARE-010

**Description**:
Implement the command to remove shares.

**Acceptance Criteria**:
- [ ] `ayo share rm <name>` removes by workspace name
- [ ] `ayo share rm <path>` removes by original host path
- [ ] `ayo share rm --all` removes all shares
- [ ] Shows success message: `✓ Unshared 'project'`
- [ ] Shows warning if share not found (not error)
- [ ] Supports `--json` output
- [ ] Symlink immediately removed from workspace

**Files**:
- `cmd/ayo/share.go`

---

### SHARE-014: Implement session share cleanup

**Priority**: P1 (High)  
**Estimate**: 1 hour  
**Dependencies**: SHARE-011

**Description**:
Ensure session shares are cleaned up when sessions end.

**Acceptance Criteria**:
- [ ] Session shares tracked with session ID
- [ ] When session ends, remove associated shares
- [ ] Hook into existing session cleanup logic
- [ ] Symlinks removed from workspace directory
- [ ] shares.json updated

**Files**:
- `internal/share/share.go`
- Session management files (TBD based on codebase)

---

## Phase 3: Integration & Testing

### SHARE-020: Integration test: share lifecycle

**Priority**: P1 (High)  
**Estimate**: 1 hour  
**Dependencies**: SHARE-011, SHARE-012, SHARE-013

**Description**:
Create integration tests for the full share lifecycle.

**Acceptance Criteria**:
- [ ] Test: share add → visible in list → visible in sandbox → rm → gone
- [ ] Test: share with --as flag
- [ ] Test: share name conflict handling
- [ ] Test: share rm by name vs by path
- [ ] Test: share rm --all
- [ ] Test: shares persist across process restart

**Files**:
- `internal/share/integration_test.go` (new) or added to existing test files

---

### SHARE-021: Integration test: sandbox visibility

**Priority**: P1 (High)  
**Estimate**: 1.5 hours  
**Dependencies**: SHARE-003, SHARE-004, SHARE-011

**Description**:
Test that shares are immediately visible inside running sandbox.

**Acceptance Criteria**:
- [ ] Start sandbox
- [ ] Create share on host
- [ ] Verify symlink exists in sandbox at `/workspace/{name}`
- [ ] Create file via sandbox
- [ ] Verify file exists on host at shared path
- [ ] Remove share
- [ ] Verify `/workspace/{name}` no longer exists in sandbox

**Files**:
- Test files TBD

---

### SHARE-022: Manual test checklist

**Priority**: P1 (High)  
**Estimate**: 30 minutes  
**Dependencies**: All SHARE-01x tickets

**Description**:
Document manual testing checklist for QA.

**Acceptance Criteria**:
- [ ] Add share system tests to MANUAL_TEST.md
- [ ] Include commands to run
- [ ] Include expected output
- [ ] Cover edge cases

**Files**:
- `MANUAL_TEST.md`

---

## Phase 4: Deprecation & Migration

### SHARE-030: Add deprecation warning to mount commands

**Priority**: P2 (Medium)  
**Estimate**: 30 minutes  
**Dependencies**: SHARE-011

**Description**:
Warn users that `ayo mount` is deprecated in favor of `ayo share`.

**Acceptance Criteria**:
- [ ] `ayo mount add` shows warning: `⚠ 'ayo mount' is deprecated, use 'ayo share' instead`
- [ ] `ayo mount list` shows same warning
- [ ] `ayo mount rm` shows same warning
- [ ] Warnings go to stderr, don't break JSON output
- [ ] Commands still work (just warn)

**Files**:
- `cmd/ayo/mount.go`

---

### SHARE-031: Implement `ayo share migrate` command

**Priority**: P2 (Medium)  
**Estimate**: 1 hour  
**Dependencies**: SHARE-011, SHARE-030

**Description**:
Provide migration path from grants to shares.

**Acceptance Criteria**:
- [ ] `ayo share migrate` reads existing mounts.json grants
- [ ] Creates equivalent shares for each grant
- [ ] Shows summary: `Migrated 3 grants to shares`
- [ ] Optionally removes old grants with `--remove-old`
- [ ] Handles name conflicts gracefully
- [ ] Dry-run mode with `--dry-run`

**Example**:
```
$ ayo share migrate --dry-run
Would migrate 3 grants:
  /Users/alex/Code/project → /workspace/project
  /Users/alex/Documents → /workspace/Documents
  /tmp/data → /workspace/data

Run without --dry-run to apply.
```

**Files**:
- `cmd/ayo/share.go`

---

### SHARE-032: Remove grant loading from sandbox creation

**Priority**: P3 (Low)  
**Estimate**: 30 minutes  
**Dependencies**: SHARE-031 (after deprecation period)

**Description**:
Once shares are stable and users migrated, remove grant loading.

**Acceptance Criteria**:
- [ ] Remove grant loading from `sandbox_manager.go`
- [ ] Remove grant loading from `pool.go`
- [ ] Remove `recreateSandboxesForMountChange()` from mount.go
- [ ] Sandboxes only have static mounts + workspace mount

**Files**:
- `internal/server/sandbox_manager.go`
- `internal/sandbox/pool.go`
- `cmd/ayo/mount.go`

**Note**: This is a future ticket, not for initial release.

---

### SHARE-033: Remove mount command entirely

**Priority**: P3 (Low)  
**Estimate**: 30 minutes  
**Dependencies**: SHARE-032, deprecation period (1-2 releases)

**Description**:
Final cleanup: remove the deprecated mount command.

**Acceptance Criteria**:
- [ ] Remove `cmd/ayo/mount.go`
- [ ] Remove mount subcommand from root.go
- [ ] Remove `internal/sandbox/mounts/` package
- [ ] Update all documentation

**Files**:
- `cmd/ayo/mount.go` (delete)
- `cmd/ayo/root.go`
- `internal/sandbox/mounts/` (delete directory)
- `AGENTS.md`

**Note**: This is a future ticket, not for initial release.

---

## Phase 5: Documentation

### SHARE-040: Update AGENTS.md with share documentation

**Priority**: P1 (High)  
**Estimate**: 30 minutes  
**Dependencies**: SHARE-011, SHARE-012, SHARE-013

**Description**:
Update the agent memory file with share system documentation.

**Acceptance Criteria**:
- [ ] Add Share Commands section
- [ ] Document `/workspace/` container path
- [ ] Update Key Files table
- [ ] Update debugging workflows
- [ ] Keep mount documentation with deprecation note

**Files**:
- `AGENTS.md`

---

### SHARE-041: Add share system to --help output

**Priority**: P1 (High)  
**Estimate**: 15 minutes  
**Dependencies**: SHARE-010

**Description**:
Ensure share commands have comprehensive help text.

**Acceptance Criteria**:
- [ ] `ayo share --help` explains the concept
- [ ] Each subcommand has examples
- [ ] Help mentions `/workspace/` container path
- [ ] Help explains instant visibility (no restart)

**Files**:
- `cmd/ayo/share.go`

---

### SHARE-042: Update README with share examples

**Priority**: P2 (Medium)  
**Estimate**: 15 minutes  
**Dependencies**: SHARE-011

**Description**:
Add share examples to project README if applicable.

**Acceptance Criteria**:
- [ ] Quick start includes `ayo share` example
- [ ] Shows workflow: share → use in sandbox → unshare

**Files**:
- `README.md` (if exists and has CLI examples)

---

## Summary

| Phase | Tickets | Estimate |
|-------|---------|----------|
| Phase 1: Core Infrastructure | SHARE-001 to SHARE-006 | ~5.5 hours |
| Phase 2: CLI Commands | SHARE-010 to SHARE-014 | ~4 hours |
| Phase 3: Integration & Testing | SHARE-020 to SHARE-022 | ~3 hours |
| Phase 4: Deprecation & Migration | SHARE-030 to SHARE-033 | ~2.5 hours |
| Phase 5: Documentation | SHARE-040 to SHARE-042 | ~1 hour |
| **Total** | **17 tickets** | **~16 hours** |

### Recommended Implementation Order

1. SHARE-002 (WorkspaceDir)
2. SHARE-001 (Share service)
3. SHARE-005 (Symlink operations)
4. SHARE-006 (Name conflicts)
5. SHARE-003 (Sandbox manager mount)
6. SHARE-004 (Pool mount)
7. SHARE-010 (CLI structure)
8. SHARE-011 (share add)
9. SHARE-012 (share list)
10. SHARE-013 (share rm)
11. SHARE-020, SHARE-021, SHARE-022 (Testing)
12. SHARE-040, SHARE-041 (Documentation)
13. SHARE-030 (Deprecation warnings)
14. SHARE-031 (Migration command)
15. SHARE-014 (Session cleanup)
16. SHARE-032, SHARE-033 (Future cleanup)
