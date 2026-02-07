---
id: ayo-dwgv
status: closed
deps: [ayo-nkig]
links: []
created: 2026-02-05T18:52:12Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, cli]
---
# Implement ayo sandbox prune

Remove all stopped sandboxes. Supports --force to skip confirmation, --all to also stop running sandboxes first.

## Acceptance Criteria

- Removes stopped sandboxes
- Confirms before pruning (unless --force)
- --all stops running sandboxes first
- Shows count of removed sandboxes

