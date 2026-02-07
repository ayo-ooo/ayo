---
id: ayo-tkj3
status: closed
deps: [ayo-15ou]
links: []
created: 2026-02-05T18:51:45Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, cli]
---
# Implement ayo sandbox show <id>

Display detailed sandbox information: ID, name, agent, image, status, created time, mounts, network config, resource limits.

## Acceptance Criteria

- Shows full sandbox details
- Displays mount points with source/dest
- Shows network config if enabled
- Handles invalid ID gracefully

