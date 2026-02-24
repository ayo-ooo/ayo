---
id: ayo-pv7g
status: open
deps: [ayo-spy5, ayo-fc4a, ayo-v7jd, ayo-m1zl, ayo-0vmu, ayo-4tpp]
links: []
created: 2026-02-24T01:02:53Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [verification, e2e, gtm]
---
# Phase 6 E2E verification - GTM Readiness

Final verification that ayo is go-to-market ready.

## Prerequisites

All Phase 6 tickets complete:
- Core docs (ayo-spy5)
- Squads/triggers docs (ayo-fc4a)
- Examples (ayo-v7jd)
- CLI help (ayo-m1zl)
- ayo doctor (ayo-0vmu)
- Test coverage (ayo-4tpp)

## Verification Checklist

### New User Onboarding (< 5 minutes)
- [ ] Install instructions work on macOS
- [ ] Install instructions work on Linux
- [ ] `ayo doctor` passes on fresh install
- [ ] First agent runs successfully
- [ ] User understands what happened

### Documentation
- [ ] README explains what ayo is
- [ ] Getting started guide is complete
- [ ] All concepts documented
- [ ] Examples are copy-pasteable
- [ ] No references to removed features

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

### Mental Model Test
A user should be able to answer:
- [ ] "What is ayo?" → CLI for managing AI agents in sandboxes
- [ ] "Where do agents run?" → In sandboxes (shared or isolated)
- [ ] "What's a squad?" → A team of agents with their own sandbox
- [ ] "What's a trigger?" → What makes an agent act without prompting
- [ ] "What's --no-jodas?" → Auto-approve mode for power users

## Acceptance Criteria

New user can install and run their first agent in under 5 minutes with clear understanding of what ayo does.
