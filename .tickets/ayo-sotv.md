---
id: ayo-sotv
status: in_progress
deps: []
links: []
created: 2026-02-12T19:47:07Z
type: chore
priority: 3
assignee: Alex Cabrera
parent: ayo-7j3v
tags: [cleanup, refactor]
---
# Consolidate duplicate formatDuration functions

3 implementations of formatDuration exist:
- cmd/ayo/flows.go:934
- internal/ui/spinner.go:256
- internal/run/print_writer.go:129

Create single shared utility and update all call sites.

## Acceptance Criteria

Single formatDuration utility, all duplicates removed, tests pass

