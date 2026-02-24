---
id: ayo-rdao
status: open
deps: [ayo-ydub]
links: []
created: 2026-02-23T22:15:03Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, cli]
---
# Remove cmd/ayo/chat.go

Delete the `chat` command for web-based chat interface.

## Context

The `ayo chat` command opened a browser-based chat UI. With `internal/server/` removed (ayo-ydub), this has no backend.

## Files to Delete

- `cmd/ayo/chat.go`

## Verification Steps

1. Delete `cmd/ayo/chat.go`
2. Run `go build ./cmd/ayo/...` - should pass
3. Run `ayo --help` - verify "chat" no longer appears

## Acceptance Criteria

- [ ] `cmd/ayo/chat.go` deleted
- [ ] Build passes
- [ ] `ayo chat` returns "unknown command"
