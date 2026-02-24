---
id: ayo-qbsu
status: open
deps: [ayo-ieiy, ayo-ydub]
links: []
created: 2026-02-23T22:15:19Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-6h19
tags: [daemon, cleanup]
---
# Clean up daemon server.go

Final cleanup of daemon after removing webhook and server dependencies.

## Context

After removing:
- Webhook server (ayo-ieiy)
- REST API server (ayo-ydub)

The daemon server.go may have orphaned code, unused imports, and dead references.

## Tasks

1. Remove unused imports
2. Remove orphaned handler registrations
3. Simplify daemon startup sequence
4. Remove any references to removed components
5. Update daemon RPC protocol if needed

## Verification Steps

1. Run `go vet ./internal/daemon/...`
2. Run `golangci-lint run ./internal/daemon/...`
3. Run `go test ./internal/daemon/...`
4. Start daemon, verify it functions
5. Test RPC communication still works

## Acceptance Criteria

- [ ] No unused imports
- [ ] No dead code warnings from linter
- [ ] Daemon starts cleanly
- [ ] All daemon tests pass
- [ ] RPC works (trigger engine, sandbox ops)
