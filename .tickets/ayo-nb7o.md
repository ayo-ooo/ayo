---
id: ayo-nb7o
status: open
deps: [ayo-dwgv]
links: []
created: 2026-02-05T18:53:07Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, testing]
---
# Integration tests for sandbox CLI

End-to-end tests for sandbox CLI commands. Tests lifecycle: create, exec, shell, push/pull, sync, stop, prune.

## Acceptance Criteria

- Tests run in CI
- Covers all sandbox subcommands
- Uses mock provider for unit tests
- Integration tests with real provider (skipped if unavailable)
- Tests error handling for each command

