---
id: ayo-ieiy
status: open
deps: []
links: []
created: 2026-02-23T22:15:11Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, daemon]
---
# Remove webhook server from daemon

Delete webhook HTTP server from daemon. Keep trigger engine (cron + file watch).

## Context

The daemon has a webhook server for receiving external HTTP triggers. This is premature - we want to focus on cron and file watch triggers first.

## Files to Modify/Delete

1. **Delete**: `internal/daemon/webhook_server.go`
2. **Delete**: `internal/daemon/webhook_server_test.go`
3. **Modify**: `internal/daemon/server.go` - remove webhook startup
4. **Modify**: `internal/daemon/protocol.go` - remove webhook-related RPC methods

## What to Keep

- `internal/daemon/trigger_engine.go` - cron and file watch triggers
- Trigger configuration loading
- Trigger execution

## Verification Steps

1. Delete webhook files
2. Remove webhook references from server.go
3. Run `go build ./...` - should pass
4. Run `go test ./internal/daemon/...` - should pass
5. Start daemon, verify it runs without webhook

## Acceptance Criteria

- [ ] Webhook files deleted
- [ ] Daemon starts without error
- [ ] Trigger engine still works (cron schedules fire)
- [ ] Tests pass
