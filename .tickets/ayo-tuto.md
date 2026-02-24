---
id: ayo-tuto
title: Tutorial guides polish
status: closed
priority: medium
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# Tutorial guides polish

Update all tutorial guides for consistency and accuracy.

## Files to Update

### docs/tutorials/first-agent.md

| Line | Issue |
|------|-------|
| 20-21 | Path uses `@reviewer/` - verify @ prefix |
| 94 | Hardcoded `claude-sonnet-4-20250514` |
| 209 | Hardcoded `claude-sonnet-4-20250514` |

### docs/tutorials/memory.md

| Line | Issue |
|------|-------|
| 271 | Hardcoded `claude-sonnet-4-20250514` |

### docs/tutorials/plugins.md

| Line | Issue |
|------|-------|
| 32 | Changed MIT → Apache-2.0 (already done) |
| 77 | Hardcoded `claude-sonnet-4-20250514` |

### docs/tutorials/triggers.md

| Line | Issue |
|------|-------|
| 56-57 | Verify `--cron` flag vs positional argument |
| 201-215 | Verify `ayo trigger` syntax |

### docs/tutorials/squads.md

| Line | Issue |
|------|-------|
| General | Verify squad commands match implementation |

## Common Changes Across Tutorials

### Model Names

Replace:
```json
"model": "claude-sonnet-4-20250514"
```

With:
```json
"model": "your-model-here"  // e.g., claude-sonnet-4-20250514
```

### Command Forms

Use singular forms consistently:
- `ayo agent` not `ayo agents`
- `ayo trigger` not `ayo triggers`

### Config Format

Use `ayo.json` format in new examples:
```json
{
  "$schema": "https://ayo.dev/schemas/agent.json",
  "agent": {
    "model": "your-model",
    "description": "..."
  }
}
```

## Acceptance Criteria

- [ ] All tutorials use placeholder model names
- [ ] All tutorials use singular command forms
- [ ] All tutorials use ayo.json format
- [ ] All tutorials commands verified against CLI
- [ ] Consistent path references

## Dependencies

- ayo-prov (model placeholders)
- ayo-cmds (command accuracy)
- ayo-path (path consistency)

## Notes

Tutorials are where users learn by doing - they must work exactly as written.
