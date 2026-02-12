---
id: ayo-cbvp
status: closed
deps: []
links: []
created: 2026-02-12T19:46:48Z
type: chore
priority: 3
assignee: Alex Cabrera
parent: ayo-7j3v
tags: [cleanup, run]
---
# Remove unused parameters in internal/run package

Remove unused function parameters flagged by gopls:
- external_tools.go:79 depth unused
- run.go:1140 ui unused

## Acceptance Criteria

gopls unusedparams warnings for internal/run reduced to zero

