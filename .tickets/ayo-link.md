---
id: ayo-link
title: Remove external link and schema references
status: closed
priority: high
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# Remove external link and schema references

Documentation references external URLs and schemas that may not exist.

## Problem

Docs reference:
1. `github.com/anthropics/ayo` - Wrong organization
2. `ayo.dev` - May not exist
3. JSON schemas at `ayo.dev/schemas/` - May not exist

## External References to Fix

### GitHub Organization

| File | Line | Issue |
|------|------|-------|
| `docs/reference/plugins.md:312` | `github.com/anthropics/ayo-plugins-devtools` |
| `docs/reference/rpc.md:1094` | `github.com/anthropics/ayo/internal/daemon` |
| `docs/advanced/extending.md:194,291,421,574` | `github.com/anthropics/ayo` |
| `docs/advanced/troubleshooting.md:665` | `github.com/anthropics/ayo/issues` |

**Fix**: Replace with `github.com/alexcabrera/ayo` or make generic

### External URLs

| File | Line | URL |
|------|------|-----|
| `docs/advanced/troubleshooting.md:667-668` | `ayo.dev/docs`, `discord.gg/ayo` |

**Fix**: Remove or mark as "Coming Soon" until these exist

### JSON Schema URLs

| File | Line | URL |
|------|------|-----|
| `docs/reference/ayo-json.md:26` | `https://ayo.dev/schemas/ayo.json` |
| `docs/reference/ayo-json.md:65` | Schema URL |
| `docs/reference/ayo-json.md:249` | Schema URL |
| `docs/reference/ayo-json.md:356-357` | Schema URL |

**Fix**: Either:
1. Remove schema URLs until ayo.dev exists
2. Host schemas on GitHub (e.g., `raw.githubusercontent.com/...`)
3. Use local schema path (`./schemas/ayo.schema.json`)

## Recommended Approach

### For GitHub References

Use generic text without linking:
```markdown
Report issues on the project's GitHub repository.
```

Or use actual repo:
```markdown
[GitHub Issues](https://github.com/alexcabrera/ayo/issues)
```

### For Schema URLs

Option A - Remove for now:
```json
{
  "model": "your-model"
}
```

Option B - Comment out:
```json
{
  // "$schema": "https://ayo.dev/schemas/ayo.json",
  "model": "your-model"
}
```

### For External URLs

Add note:
```markdown
> **Community**: Documentation site and Discord coming soon.
```

## Acceptance Criteria

- [ ] No references to `github.com/anthropics/`
- [ ] No broken external links (`ayo.dev`, `discord.gg`)
- [ ] Schema URLs either removed or point to valid location
- [ ] All GitHub links point to correct repository

## Dependencies

None

## Notes

This is a release blocker - broken links are embarrassing.
