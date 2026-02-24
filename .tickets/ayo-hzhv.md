---
id: ayo-hzhv
status: open
deps: [ayo-7dui, ayo-la11, ayo-nqyv, ayo-7jth, ayo-mp44]
links: []
created: 2026-02-24T01:02:53Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-pv3a
tags: [verification, e2e]
---
# Phase 3 E2E verification

End-to-end verification of unified ayo.json schema.

## Prerequisites

All Phase 3 tickets complete:
- Agent schema (ayo-7dui)
- Agent loader (ayo-la11)
- Squad schema (ayo-nqyv)
- Squad loader (ayo-7jth)
- SQUAD.md migration (ayo-mp44)

## Verification Checklist

### Agent Config (ayo.json)
- [ ] Agent with ayo.json loads correctly
- [ ] Legacy config.json still works (with deprecation warning)
- [ ] All ayo.json fields respected (model, tools, permissions, etc)
- [ ] JSON schema validates configs
- [ ] Invalid config gives clear error message

### Squad Config (ayo.json)
- [ ] Squad with ayo.json loads correctly
- [ ] Legacy SQUAD.md frontmatter still works (with warning)
- [ ] All squad fields respected (lead, agents, planners, etc)
- [ ] SQUAD.md body still used for constitution

### Migration
- [ ] `ayo migrate` command works
- [ ] Converts SQUAD.md frontmatter to ayo.json
- [ ] Strips frontmatter from SQUAD.md
- [ ] Preserves all configuration values

### Schema Validation
- [ ] `schemas/ayo.json` exists
- [ ] Editor autocompletion works (VSCode)
- [ ] Invalid fields caught by schema
- [ ] Required fields enforced

## Acceptance Criteria

Both agents and squads use ayo.json seamlessly. Migration path is clear.
