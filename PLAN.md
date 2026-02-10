# Ayo Share System: Comprehensive Plan

## Summary

Replace the complex mount/grants system with a simpler `ayo share` command that symlinks user paths into a pre-mounted workspace. This eliminates sandbox restarts when sharing changes.

## Current State (Problems)

### Existing Architecture

```
Host                              Container
────────────────────              ────────────────────
~/.local/share/ayo/sandbox/
  homes/                    →     /home/{agent}/
  shared/                   →     /shared/
  workspaces/               →     /workspaces/{session}/
/tmp/ayo                    →     /run/ayo/

(Dynamic per grant)
/Users/alex/Code/project    →     /Users/alex/Code/project
```

### Problems with Current Approach

1. **Sandbox restart required**: `ayo mount add` saves a grant, but the sandbox needs recreation to apply the new mount
2. **Three overlapping layers**: Grants, project mounts (.ayo.json), session mounts (--mount) that can only restrict, not grant
3. **Confusing mental model**: Mounts ≠ grants ≠ permissions
4. **Performance hit**: Every mount change → stop sandbox → restart daemon → wait for new sandbox

## Proposed Architecture

### Core Concept

Pre-mount a **workspace directory** at sandbox creation. User shares are **symlinks inside that workspace**, not new container mounts.

```
Host                                    Container
────────────────────                    ────────────────────
~/.local/share/ayo/sandbox/
  homes/                          →     /home/{agent}/
  shared/                         →     /shared/         (agent-to-agent)
  workspaces/                     →     /workspaces/     (session data)
  workspace/                      →     /workspace/      (NEW: user shares)
    project/      → symlink to ~/Code/project
    docs/         → symlink to ~/Documents/specs
```

### Directory Purposes

| Container Path | Purpose | Managed By |
|---------------|---------|------------|
| `/home/{agent}/` | Agent private storage | System (per-agent) |
| `/shared/` | Cross-agent collaboration | System (sticky bit) |
| `/workspaces/{session}/` | Session scratch space | System (ephemeral) |
| `/workspace/` | **User-shared host files** | User (`ayo share`) |

### Why This Works

1. **`/workspace/` is mounted once** at sandbox creation (empty or with existing shares)
2. **Symlinks are filesystem operations** inside a mounted directory — no container restart needed
3. **Everything in `/workspace/` is r/w** — safety comes from the working copy model, not permissions
4. **Simple mental model**: "Share a folder" = "It appears in /workspace/"

## New CLI: `ayo share`

### Commands

```bash
# Share a path (creates symlink in workspace)
ayo share /path/to/project           # → /workspace/project/
ayo share ~/Documents/specs          # → /workspace/specs/
ayo share .                           # → /workspace/{current-dir-name}/

# Share with custom name
ayo share /path/to/project --as myproj   # → /workspace/myproj/

# Session-only share (removed when session ends)
ayo share /path --session

# List shares
ayo share list
ayo share list --json

# Unshare
ayo share rm project                  # By workspace name
ayo share rm /path/to/project         # By original path
ayo share rm --all                    # Remove all shares
```

### Behavior

1. `ayo share /path`:
   - Resolve path to absolute
   - Validate path exists
   - Create symlink: `~/.local/share/ayo/sandbox/workspace/{name} → /path`
   - Symlink immediately visible in running sandbox at `/workspace/{name}`

2. `ayo share rm`:
   - Remove symlink from workspace directory
   - Immediately reflected in sandbox

3. No daemon restart, no sandbox recreation.

## Data Model

### shares.json

```json
{
  "version": 1,
  "shares": [
    {
      "name": "project",
      "path": "/Users/alex/Code/project",
      "session": false,
      "shared_at": "2025-02-09T12:00:00Z"
    },
    {
      "name": "specs",
      "path": "/Users/alex/Documents/specs", 
      "session": false,
      "shared_at": "2025-02-09T12:00:00Z"
    }
  ]
}
```

Location: `~/.local/share/ayo/shares.json` (production) or `.local/share/ayo/shares.json` (dev mode)

### Persistence

- **Permanent shares**: Survive sandbox restart, stored in shares.json
- **Session shares** (`--session`): Removed when session ends, tracked in memory or session db

## Implementation Plan

### Phase 1: Core Infrastructure

1. **Create new package**: `internal/share/`
   - `share.go`: Share service (load, save, add, remove, list)
   - Types: `Share`, `SharesFile`

2. **Create workspace directory at init**:
   - In `internal/sync/git.go`: Add `WorkspaceDir()` function
   - Create `~/.local/share/ayo/sandbox/workspace/` alongside homes/shared

3. **Mount workspace directory**:
   - In `internal/server/sandbox_manager.go`: Add mount for workspace dir → `/workspace/`
   - In `internal/sandbox/pool.go`: Same for pool-managed sandboxes

### Phase 2: CLI Commands

1. **Create `cmd/ayo/share.go`**:
   - `ayo share <path>` - Add share
   - `ayo share list` - List shares
   - `ayo share rm <name|path>` - Remove share
   - Flags: `--as`, `--session`, `--json`

2. **Symlink operations**:
   - Share add: `os.Symlink(absPath, workspaceDir/name)`
   - Share rm: `os.Remove(workspaceDir/name)`
   - No sandbox interaction needed

### Phase 3: Deprecate Grants System

1. **Keep `ayo mount` temporarily** for backward compatibility
2. **Show deprecation warning**: "Use 'ayo share' instead"
3. **Migration command**: `ayo share migrate` converts grants to shares
4. **Remove in future version**

### Phase 4: Cleanup

1. Remove `internal/sandbox/mounts/grants.go` (after deprecation period)
2. Remove grant-loading code from pool.go and sandbox_manager.go
3. Remove `recreateSandboxesForMountChange()` from mount.go
4. Remove .ayo.json project mount handling (or repurpose for other config)

## Edge Cases & Considerations

### Symlink vs Bind Mount

**Symlinks work because**:
- Parent directory (`/workspace/`) is already mounted
- Symlinks are resolved inside the container, pointing to paths that exist via the mount
- No nested mount support needed

**Potential issue**: If user shares `/foo/bar` and later shares `/foo`, the symlinks work independently (both point to host paths).

### Name Conflicts

- Auto-generated names use basename: `~/Code/project` → `project`
- If `project` exists, fail with error and suggest `--as different-name`
- Or auto-suffix: `project-2`

### Absolute vs Relative Symlinks

Use **absolute symlinks** pointing to host paths:
```
/workspace/project → /Users/alex/Code/project
```

Inside container, this resolves correctly because the parent mount makes the host path accessible.

**Wait, that's not right.** The symlink target is a *host* path, but inside the container there's no `/Users/alex/...`. 

**Correction**: We need the workspace directory itself to be on the host, with symlinks created on the host. When mounted into the container, the symlinks resolve on the host side (virtiofs/bind mount follows symlinks).

```
Host:
  ~/.local/share/ayo/sandbox/workspace/
    project → /Users/alex/Code/project  (symlink)

Container sees (via mount):
  /workspace/
    project/  (contents of /Users/alex/Code/project)
```

This works because bind mounts follow symlinks by default.

### Dev Mode

Dev mode affects paths but not behavior:
- Production: `~/.local/share/ayo/sandbox/workspace/`
- Dev mode: `.local/share/ayo/sandbox/workspace/`

The `paths.DataDir()` already handles this; WorkspaceDir() uses it.

### Working Copy Model

The working copy system (`/workspaces/{session}/`) is **separate** from `/workspace/`:
- `/workspace/` = direct access to shared host files
- `/workspaces/{session}/` = copied files for safe agent manipulation

If we want file-level safety:
1. Agent requests file from `/workspace/project/foo.go`
2. System copies to `/workspaces/{session}/scratch/foo.go`
3. Agent works on copy
4. User syncs back with `ayo sandbox sync`

**Decision**: Keep both systems. `/workspace/` for direct access, `/workspaces/` for copy-based safety when needed.

## Files to Modify

| File | Changes |
|------|---------|
| `internal/share/share.go` | **NEW**: Share service |
| `internal/sync/git.go` | Add `WorkspaceDir()`, create dir in `Init()` |
| `internal/server/sandbox_manager.go` | Add workspace mount |
| `internal/sandbox/pool.go` | Add workspace mount |
| `cmd/ayo/share.go` | **NEW**: CLI commands |
| `cmd/ayo/root.go` | Add share subcommand |
| `cmd/ayo/mount.go` | Add deprecation warnings |
| `AGENTS.md` | Update documentation |

## Testing Checklist

- [ ] `ayo share /tmp/test` creates symlink in workspace dir
- [ ] Symlink immediately visible inside running sandbox at `/workspace/test`
- [ ] Can read/write files through the share
- [ ] `ayo share rm test` removes symlink
- [ ] Removal immediately reflected in sandbox
- [ ] `ayo share list` shows all shares
- [ ] `--as` flag works for custom naming
- [ ] `--session` flag creates non-persistent share
- [ ] Shares survive sandbox restart
- [ ] Dev mode uses correct paths
- [ ] Deprecation warning shown for `ayo mount` commands

## Open Questions (Resolved)

1. **Symlink vs bind mount?** → Symlinks on host, followed by bind mount
2. **Persistence?** → Yes, shares.json
3. **What about existing directories?** → `/workspace/` for human↔agent sharing, keep `/shared/` for agent↔agent
4. **Session shares?** → Yes with `--session` flag, not default

## Migration Path

1. Release with both `ayo mount` and `ayo share`
2. `ayo mount` shows deprecation warning
3. `ayo share migrate` converts existing grants to shares
4. Future release removes `ayo mount`

## Summary

| Before | After |
|--------|-------|
| `ayo mount add` → restart sandbox | `ayo share` → instant |
| Grants + project mounts + CLI mounts | Just shares |
| 3 permission layers | r/w for everything shared |
| Safety via permissions | Safety via working copy |
| Complex mental model | "Share = visible in /workspace/" |
