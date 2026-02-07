---
id: ayo-12vf
status: closed
deps: [ayo-15ou]
links: []
created: 2026-02-05T18:51:37Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, cli]
---
# Implement ayo sandbox list

List active sandboxes with columns: ID, Agent, Status, Created, Mounts. Uses SandboxProvider.List() from providers interface.

## Acceptance Criteria

- Shows table of running sandboxes
- Includes sandbox ID, agent handle, status, creation time
- Works with apple-container provider
- Empty state message when no sandboxes

