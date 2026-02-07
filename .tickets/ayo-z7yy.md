---
id: ayo-z7yy
status: open
deps: [ayo-1rw2]
links: []
created: 2026-02-05T18:52:38Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, files, workspace]
---
# Working Copy Model: isolated workspace with sync

Sandbox uses copy-on-write workspace. Project dir mounted read-only, changes go to overlay. sync command merges changes back to host with diff preview and user approval.

## Acceptance Criteria

- Project mounted read-only by default
- Agent writes go to overlay layer
- ayo sandbox sync <id> shows diff of changes
- User approves which changes to apply
- Applied changes written to host filesystem
- --dry-run shows what would sync

