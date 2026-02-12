---
id: ayo-33qq
status: open
deps: []
links: []
created: 2026-02-12T19:47:33Z
type: chore
priority: 4
assignee: Alex Cabrera
parent: ayo-7j3v
tags: [cleanup, skills]
---
# Remove deprecated UserSharedDir field

internal/skills/discover.go:17-18 has UserSharedDir marked deprecated but still used. Either complete the migration away from it or remove the deprecation marker.

## Acceptance Criteria

UserSharedDir either removed or deprecation marker removed

