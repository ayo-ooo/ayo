---
id: ayo-refs
title: Reference documents polish
status: open
priority: medium
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# Reference documents polish

Update all reference documents for accuracy and completeness.

## Files to Update

### docs/reference/ayo-json.md

| Line | Issue |
|------|-------|
| 26 | Schema URL may not exist |
| 27 | Provider hardcoded as "anthropic" |
| 28, 53, 69 | Hardcoded model names |
| 65 | Schema URL may not exist |
| 249 | Schema URL may not exist |
| 356-357 | Schema URL may not exist |

**Changes needed:**
- Use placeholder provider/model
- Remove or update schema URLs
- Add all provider options

### docs/reference/cli.md

| Line | Issue |
|------|-------|
| 26 | `ayo service` should be `ayo sandbox service` |
| 70-100 | Mix of `ayo agents` and `ayo agent` |
| 588-596 | Incomplete environment variables |

**Changes needed:**
- Fix service command path
- Standardize on singular command forms
- Complete environment variable section
- Add missing commands (see ayo-feat ticket)

### docs/reference/plugins.md

| Line | Issue |
|------|-------|
| 312 | `github.com/anthropics/ayo-plugins-devtools` |
| 59, 313 | Changed MIT → Apache-2.0 (done) |

**Changes needed:**
- Fix GitHub repository URL

### docs/reference/rpc.md

| Line | Issue |
|------|-------|
| 1094 | `github.com/anthropics/ayo/internal/daemon` |

**Changes needed:**
- Fix import path to actual repository

### docs/reference/prompts.md

| Issue |
|-------|
| Verify prompt paths match implementation |

## Acceptance Criteria

- [ ] ayo-json.md uses placeholders
- [ ] ayo-json.md schema URLs resolved
- [ ] cli.md complete and accurate
- [ ] plugins.md correct repository
- [ ] rpc.md correct import paths
- [ ] All references verified

## Dependencies

- ayo-prov (provider neutrality)
- ayo-cmds (command accuracy)
- ayo-link (URL fixes)
- ayo-env (environment variables)
- ayo-feat (missing commands)

## Notes

Reference docs must be authoritative - every example must work.
