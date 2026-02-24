---
id: ayo-evik
status: open
deps: [ayo-dicu]
links: []
created: 2026-02-23T23:13:32Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-whmn
tags: [cli, permissions]
---
# Implement --no-jodas CLI flag

Add `--no-jodas` flag for auto-approving file_request operations.

## Flag Definition

```go
// cmd/ayo/root.go
rootCmd.PersistentFlags().Bool("no-jodas", false, 
    "Auto-approve all file modifications (use with caution)")
```

## Usage

```bash
# Single command
ayo --no-jodas "refactor my entire codebase"

# Shortened alias
ayo -y "do the thing"  # -y is alias for --no-jodas
```

## Implementation

### Files to Modify

1. `cmd/ayo/root.go`
   - Add `--no-jodas` and `-y` flags
   - Store in viper config for access by tools
   - Add to cobra command

2. `internal/tools/file_request.go`
   - Check for no-jodas mode before prompting
   - If enabled, auto-approve and log

3. `cmd/ayo/run.go` (or wherever agent execution starts)
   - Pass no-jodas state to tool context

### Approval Priority

Check in order:
1. Session cache ("Always approve" was selected earlier)
2. `--no-jodas` CLI flag
3. Agent-level `permissions.auto_approve` in ayo.json
4. Global `permissions.no_jodas` in config.json
5. Prompt user

### Help Text

```
Flags:
  -y, --no-jodas   Auto-approve all file modifications without prompting.
                   WARNING: Agent can modify any file in your home directory.
                   All modifications are logged to ~/.local/share/ayo/audit.log
```

## Safety

Even with --no-jodas:
- Modifications are limited to `/mnt/{user}/` boundary
- All changes logged to audit.log (see ayo-vclt)
- Can combine with `--dry-run` to preview changes

## Testing

- Test flag is parsed correctly
- Test precedence over other approval methods
- Test audit logging still works
- Test help text displays correctly
