---
id: ayo-2xh8
status: closed
deps: [ayo-15ou]
links: []
created: 2026-02-05T18:52:03Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, cli]
---
# Implement ayo sandbox logs <id>

View sandbox container logs. Supports --follow for streaming, --tail for last N lines.

## Acceptance Criteria

- Shows container logs
- --follow streams new logs
- --tail N shows last N lines
- Handles containers with no logs

