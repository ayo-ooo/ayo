---
id: ayo-pv7g
status: open
deps: [ayo-m1zl, ayo-0vmu, ayo-4tpp, ayo-2e0t]
links: []
created: 2026-02-24T01:02:53Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [verification, e2e, phase7]
---
# Phase 7 E2E Verification - CLI Polish

End-to-end verification that Phase 7 CLI polish is complete.

## Prerequisites

All Phase 7 tickets complete:
- CLI help polish (ayo-m1zl)
- ayo doctor improvements (ayo-0vmu)
- Test coverage (ayo-4tpp)
- AGENTS.md update (ayo-2e0t)

## Verification Checklist

### CLI Help
- [ ] `ayo --help` is informative
- [ ] All subcommands have help
- [ ] Examples in help text
- [ ] No typos or outdated info

### ayo doctor
- [ ] Catches missing sandbox provider
- [ ] Catches daemon not running
- [ ] Catches config issues
- [ ] Gives actionable fix suggestions
- [ ] Green checkmarks for passing checks

### Test Coverage
- [ ] `go test ./...` passes
- [ ] Coverage > 70%
- [ ] Critical paths tested
- [ ] Integration tests pass

### AGENTS.md
- [ ] Memory file is accurate
- [ ] Reflects current codebase
- [ ] Key commands documented

## Acceptance Criteria

CLI is polished, tests pass, and codebase is ready for final documentation phase.
