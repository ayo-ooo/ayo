---
id: ayo-guid
title: Guide documents polish
status: open
priority: medium
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# Guide documents polish

Update all guide documents for consistency and accuracy.

## Files to Update

### docs/guides/agents.md

| Line | Issue |
|------|-------|
| 21 | Hardcoded `claude-sonnet-4-20250514` |
| 66 | Hardcoded `claude-sonnet-4-20250514` |
| 330 | Hardcoded `claude-sonnet-4-20250514` |

### docs/guides/squads.md

| Line | Issue |
|------|-------|
| 112-117 | Hardcoded `claude-sonnet-4-20250514` |

### docs/guides/triggers.md

| Line | Issue |
|------|-------|
| 22-57 | Verify trigger syntax (--cron vs positional) |
| 63-68 | Verify show/enable/disable/remove syntax |

### docs/guides/sandbox.md

| Issue |
|-------|
| Verify sandbox commands match `ayo sandbox --help` |

### docs/guides/security.md

| Issue |
|-------|
| Verify trust levels match implementation |
| Check guardrails documentation |

### docs/guides/tools.md

| Issue |
|-------|
| Verify tool configuration syntax |

## Common Changes

### Model Names

Use placeholders with comments:
```json
"model": "your-model"  // e.g., claude-sonnet-4-20250514, gpt-4o
```

### Command Examples

Verify against actual CLI output:
```bash
go run ./cmd/ayo/... [command] --help
```

## Acceptance Criteria

- [ ] All model names use placeholders
- [ ] All commands verified
- [ ] Sandbox guide matches current implementation
- [ ] Security guide trust levels accurate
- [ ] Tools guide configuration accurate

## Dependencies

- ayo-prov (model placeholders)
- ayo-cmds (command accuracy)

## Notes

Guides are for deeper understanding - they can be more detailed but must be accurate.
