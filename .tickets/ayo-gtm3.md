---
id: ayo-gtm3
status: deferred
deps: []
links: []
created: 2026-02-24T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-gtm1
tags: [gtm, refactoring]
---
# Task: Split internal/daemon/server.go

## Summary

`internal/daemon/server.go` is 1170 lines. Split into logical modules for maintainability.

## Current Structure Analysis

The file contains:
1. **Server struct** - Core server definition, lifecycle (~150 lines)
2. **RPC handlers** - General RPC method implementations (~300 lines)
3. **Squad RPC** - Squad-specific RPC handlers (~250 lines)
4. **Ticket RPC** - Ticket-specific RPC handlers (~200 lines)
5. **Session management** - Session lifecycle (~150 lines)
6. **Health/Status** - Health checks, status reporting (~120 lines)

## Proposed Split

| New File | Contents | Est. Lines |
|----------|----------|------------|
| `server.go` | Core Server struct, Start/Stop, lifecycle | ~200 |
| `rpc_handlers.go` | General RPC methods (exec, file ops) | ~300 |
| `squad_handlers.go` | Squad-specific RPC (SquadCreate, etc.) | ~250 |
| `ticket_handlers.go` | Ticket RPC handlers | ~200 |
| `session_handlers.go` | Session management RPC | ~150 |
| `status.go` | Health checks, status reporting | ~120 |

Note: `squad_rpc.go` and `ticket_rpc.go` may already exist - verify and merge.

## Implementation Steps

1. [ ] Audit existing files in daemon/ to avoid duplication
2. [ ] Create `rpc_handlers.go` - move general RPC methods
3. [ ] Verify/update `squad_handlers.go` 
4. [ ] Create `ticket_handlers.go` - move ticket RPC
5. [ ] Create `session_handlers.go` - move session management
6. [ ] Create `status.go` - move health/status code
7. [ ] Update imports
8. [ ] Run tests

## Acceptance Criteria

- [ ] server.go < 300 lines
- [ ] No handler file > 350 lines
- [ ] All tests pass
- [ ] Clean separation of concerns
