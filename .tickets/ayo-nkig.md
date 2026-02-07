---
id: ayo-nkig
status: closed
deps: [ayo-15ou]
links: []
created: 2026-02-05T18:52:07Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, cli]
---
# Implement ayo sandbox stop <id>

Stop a running sandbox. Supports --force for immediate termination, --timeout for graceful shutdown period.

## Acceptance Criteria

- Stops sandbox gracefully by default
- --force kills immediately
- --timeout sets shutdown wait time
- Confirms sandbox stopped

