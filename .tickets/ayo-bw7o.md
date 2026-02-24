---
id: ayo-bw7o
status: closed
deps: [ayo-evik]
links: []
created: 2026-02-23T23:13:32Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-whmn
tags: [config, permissions]
---
# Add no_jodas to global config

Add `permissions.no_jodas` field to global config at `~/.config/ayo/config.json`. When true, acts as if `--no-jodas` was always passed, auto-approving all file_request operations.

## Context

Users who fully trust their agents shouldn't need to pass `--no-jodas` on every command. This ticket adds a global config option for persistent no-jodas mode.

This is the lowest priority in the approval chain (ayo-evik):
1. Session cache
2. `--no-jodas` CLI flag
3. Agent-level `permissions.auto_approve`
4. **Global `permissions.no_jodas` in config.json** ← This ticket
5. Prompt user

## Config Schema

```json
// ~/.config/ayo/config.json
{
  "permissions": {
    "no_jodas": true
  }
}
```

## Files to Modify

1. **`internal/config/config.go`** - Add `Permissions.NoJodas` field
2. **`internal/tools/file_request.go`** - Check global config in approval chain
3. **`docs/configuration.md`** - Document the new field

## Implementation

```go
// internal/config/config.go
type Permissions struct {
    NoJodas bool `json:"no_jodas,omitempty"`
}

type Config struct {
    // ... existing fields ...
    Permissions Permissions `json:"permissions,omitempty"`
}
```

## Acceptance Criteria

- [ ] `permissions.no_jodas` field parses from global config
- [ ] Field is optional and defaults to false
- [ ] file_request checks this field in approval chain
- [ ] Setting is lowest priority (all other methods override it)
- [ ] Audit logging still works when auto-approved via this setting
- [ ] Documentation updated with example and warning

## Warning in Docs

```markdown
**⚠️ Warning**: Setting `permissions.no_jodas: true` globally means all agents
can modify any file in your home directory without prompting. All modifications
are logged to `~/.local/share/ayo/audit.log`, but use with caution.
```

## Testing

- Test parsing valid JSON with permissions.no_jodas: true
- Test parsing JSON without permissions (should default false)
- Test file_request respects this setting
- Test priority: agent auto_approve overrides this setting
