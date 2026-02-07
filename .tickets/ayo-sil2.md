---
id: ayo-sil2
status: closed
deps: [ayo-cton]
links: []
created: 2026-02-05T18:51:58Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, cli]
---
# Implement ayo sandbox shell <id>

Open interactive shell in sandbox. Falls back to non-TTY mode with readline if TTY not supported. Supports --as @agent flag to shell as agent user.

## Acceptance Criteria

- Opens shell in sandbox
- Works without TTY (Apple Container limitation)
- --as flag sets user identity
- Shows sandbox info on connect
- Clean exit handling

