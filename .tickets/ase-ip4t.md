---
id: ase-ip4t
status: closed
deps: [ase-o8c9, ase-cy5l]
links: []
created: 2026-02-09T03:25:24Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-fked
---
# Update ayo agents CLI with trust levels and metadata

## Background

The `ayo agents` CLI needs updates to display new metadata: trust levels, creation source, capabilities summary.

## Current State

```bash
$ ayo agents list
@ayo        builtin
@researcher user
```

## Target State

```bash
$ ayo agents list
NAME           TYPE      TRUST        CREATED BY
@ayo           builtin   sandboxed    system
@researcher    user      sandboxed    user
@doc-writer    user      sandboxed    @ayo
@my-unsafe     user      ⚠ unrestricted  user

$ ayo agents show @doc-writer
Name: @doc-writer
Type: user
Trust Level: sandboxed
Created By: @ayo
Created At: 2026-02-08T20:30:00Z
Refinements: 2 (last: 2026-02-09T10:00:00Z)

System Prompt:
  You are a technical documentation writer...

Capabilities (inferred):
  - documentation-writing (0.95)
  - technical-writing (0.88)
  - markdown-formatting (0.75)

Skills:
  - file_read
  - file_write
```

## Implementation Details

### Changes to `ayo agents list`

1. Add columns: TRUST, CREATED BY
2. Use color/emoji for trust levels:
   - sandboxed: normal
   - privileged: yellow
   - unrestricted: red with ⚠

3. Add filters:
   ```bash
   ayo agents list --trust=sandboxed
   ayo agents list --created-by=@ayo
   ayo agents list --type=user
   ```

### Changes to `ayo agents show`

1. Add trust level display
2. Add creation metadata (by, at)
3. Add refinement count and last date
4. Add inferred capabilities summary
5. Format system prompt nicely

### JSON Output

```json
{
  "name": "@doc-writer",
  "type": "user",
  "trust_level": "sandboxed",
  "created_by": "@ayo",
  "created_at": "2026-02-08T20:30:00Z",
  "refinement_count": 2,
  "last_refined": "2026-02-09T10:00:00Z",
  "system_prompt": "...",
  "capabilities": [
    {"name": "documentation-writing", "confidence": 0.95}
  ],
  "skills": ["file_read", "file_write"]
}
```

### Files to Modify

1. `cmd/ayo/agents.go` - Update list and show commands
2. Add queries to join with ayo_created_agents and agent_capabilities tables

## Acceptance Criteria

- [ ] ayo agents list shows trust level column
- [ ] ayo agents list shows created by column
- [ ] Trust levels color-coded appropriately
- [ ] ayo agents show displays all new metadata
- [ ] Refinement history shown for @ayo-created agents
- [ ] Capabilities summary shown
- [ ] --json output includes all fields
- [ ] Filter flags work (--trust, --created-by)

