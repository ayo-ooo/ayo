---
id: ayo-qkp9
status: closed
deps: []
links: []
created: 2026-02-12T19:32:07Z
type: chore
priority: 2
assignee: Alex Cabrera
parent: ayo-qgnv
tags: [cleanup]
---
# Remove custom min() function, use Go 1.21 builtin

internal/sync/git.go:511 has a custom min() function that can be replaced with Go builtin

