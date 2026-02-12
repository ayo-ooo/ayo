---
id: ayo-jvjm
status: closed
deps: []
links: []
created: 2026-02-12T19:46:53Z
type: chore
priority: 3
assignee: Alex Cabrera
parent: ayo-7j3v
tags: [cleanup, ui]
---
# Remove unused parameters in internal/skills and internal/ui packages

Remove unused function parameters flagged by gopls:
- skills/skills.go:195 path unused
- ui/json_render.go:42,106 depth unused

## Acceptance Criteria

gopls unusedparams warnings for these packages reduced to zero

