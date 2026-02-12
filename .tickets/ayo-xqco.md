---
id: ayo-xqco
status: open
deps: []
links: []
created: 2026-02-12T19:46:44Z
type: chore
priority: 3
assignee: Alex Cabrera
parent: ayo-7j3v
tags: [cleanup, daemon]
---
# Remove unused parameters in internal/daemon package

Remove unused function parameters flagged by gopls:
- client.go:81 ctx unused
- conduit.go:363 ctx unused
- matrix_rpc.go:58 ctx unused
- matrix_rpc.go:245 ctx unused

## Acceptance Criteria

gopls unusedparams warnings for internal/daemon reduced to zero

