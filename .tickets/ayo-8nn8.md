---
id: ayo-8nn8
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
# Remove cmd/ayo/serve.go

Delete the `serve` command that starts the REST API server.

## Context

The `ayo serve` command started an HTTP server for web/mobile clients. With `internal/server/` removed (ayo-ydub), this command has no implementation.

## Files to Delete

- `cmd/ayo/serve.go`

## Verification Steps

1. Delete `cmd/ayo/serve.go`
2. Run `go build ./cmd/ayo/...` - should pass
3. Run `ayo --help` - verify "serve" no longer appears
4. Run `go test ./...` - should pass

## Acceptance Criteria

- [ ] `cmd/ayo/serve.go` deleted
- [ ] Build passes
- [ ] `ayo serve` returns "unknown command"
