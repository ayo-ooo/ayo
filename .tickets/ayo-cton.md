---
id: ayo-cton
status: closed
deps: [ayo-15ou]
links: []
created: 2026-02-05T18:51:53Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, cli]
---
# Implement ayo sandbox exec <id> <cmd>

Execute a command inside a sandbox. Supports --user flag to run as specific user, --workdir to set working directory. Streams stdout/stderr.

## Acceptance Criteria

- Executes command in sandbox
- Returns exit code from command
- Supports --user flag for user identity
- Supports --workdir flag
- Streams output in real-time

