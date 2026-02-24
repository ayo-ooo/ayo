---
id: ayo-trou
title: Troubleshooting guide accuracy
status: open
priority: medium
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# Troubleshooting guide accuracy

Fix inaccuracies in the advanced troubleshooting guide.

## Issues Found

### docs/advanced/troubleshooting.md

| Line | Issue |
|------|-------|
| 85-93 | Uses `ayo daemon start` (doesn't exist) |
| 113-115 | Uses `ayo daemon status` (doesn't exist) |
| 275 | Only checks `ANTHROPIC_API_KEY` |
| 617 | Hardcoded `claude-3-haiku-20240307` |
| 665 | Links to `github.com/anthropics/ayo/issues` |
| 667-668 | Links to `ayo.dev/docs`, `discord.gg/ayo` |

### Command Fixes

**Before:**
```bash
ayo daemon start
ayo daemon status
```

**After:**
```bash
ayo sandbox service start
ayo sandbox service status
```

### Provider Check Fix

**Before:**
```bash
# Check API key
echo $ANTHROPIC_API_KEY
```

**After:**
```bash
# Check configured providers
ayo doctor

# Or check environment
env | grep -E "(ANTHROPIC|OPENAI|GEMINI|OPENROUTER)_API_KEY"
```

### Model Name Fix

**Before:**
```json
"model": "claude-3-haiku-20240307"
```

**After:**
```json
"model": "your-model"  // fast model for testing
```

### External Link Fixes

**Before:**
```markdown
- [GitHub Issues](https://github.com/anthropics/ayo/issues)
- [Documentation](https://ayo.dev/docs)
- [Discord](https://discord.gg/ayo)
```

**After:**
```markdown
- [GitHub Issues](https://github.com/alexcabrera/ayo/issues)

> **Note**: Documentation site and community Discord coming soon.
```

## Acceptance Criteria

- [ ] All daemon commands → sandbox service
- [ ] Provider checks are provider-neutral
- [ ] Model names use placeholders
- [ ] GitHub links point to correct repo
- [ ] External URLs handled appropriately

## Dependencies

- ayo-cmds (command accuracy)
- ayo-prov (provider neutrality)
- ayo-link (URL fixes)

## Notes

Users in trouble need accurate guidance most of all.
