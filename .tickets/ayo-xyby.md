---
id: ayo-xyby
status: closed
deps: [ayo-evik]
links: []
created: 2026-02-23T23:13:32Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-whmn
tags: [agents, permissions]
---
# Add auto_approve to agent ayo.json

Add `permissions.auto_approve` field to the agent ayo.json schema. This allows per-agent override for file approval, enabling trusted agents to bypass file_request prompts.

## Context

The file_request approval flow (ayo-dicu) prompts users before any host filesystem modification. However, some agents are trusted (e.g., personal automation agents) and prompting every time degrades UX. This ticket adds agent-level auto-approve configuration.

This is part of the approval priority chain defined in ayo-evik:
1. Session cache ("Always approve" was selected earlier)
2. `--no-jodas` CLI flag
3. **Agent-level `permissions.auto_approve` in ayo.json** ← This ticket
4. Global `permissions.no_jodas` in config.json
5. Prompt user

## Schema Addition

```json
{
  "permissions": {
    "auto_approve": true
  }
}
```

## Files to Modify

1. **`internal/agent/config.go`** - Add `Permissions.AutoApprove` field to agent config struct
2. **`internal/tools/file_request.go`** - Check agent config before prompting
3. **`docs/agents.md`** - Document the new field

## Implementation

```go
// internal/agent/config.go
type Permissions struct {
    AutoApprove bool `json:"auto_approve,omitempty"`
}

type AgentConfig struct {
    // ... existing fields ...
    Permissions Permissions `json:"permissions,omitempty"`
}
```

In file_request handler:
```go
// Check approval chain
if session.HasApproval(path) || 
   ctx.NoJodas || 
   agent.Config.Permissions.AutoApprove ||  // <-- This ticket
   globalConfig.Permissions.NoJodas {
    return approve()
}
return promptUser()
```

## Acceptance Criteria

- [ ] `permissions.auto_approve` field parses from agent ayo.json
- [ ] Field is optional and defaults to false
- [ ] file_request checks this field in approval chain
- [ ] Setting is respected in correct priority order
- [ ] Audit logging still works when auto-approved via this setting
- [ ] Documentation updated with example

## Testing

- Test parsing valid JSON with permissions.auto_approve: true
- Test parsing JSON without permissions (should default false)
- Test file_request respects this setting
- Test priority: CLI flag overrides this setting
