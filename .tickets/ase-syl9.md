---
id: ase-syl9
status: open
deps: [ase-mwdy]
links: []
created: 2026-02-09T03:05:15Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-qnjh
---
# Implement ayo chat CLI commands

Add `ayo chat` subcommands for agents to communicate via Matrix from within sandboxes.

## Background

Agents don't have direct Matrix connections. They use the `ayo chat` CLI which communicates through the daemon. This works in sandboxes via the mounted daemon socket.

## Commands

```bash
# Room management
ayo chat rooms                          # List all rooms
ayo chat rooms --session abc123         # Filter by session
ayo chat create --name '#session-xyz'   # Create room

# Messaging  
ayo chat send '#room' 'message'         # Send message
ayo chat send '#room' --file data.json  # Send file/JSON
ayo chat send '#room' --as @agent       # Send as specific agent

# Reading
ayo chat read '#room'                   # Read recent messages (default 20)
ayo chat read '#room' --limit 50        # More messages
ayo chat read '#room' --since 1h        # Last hour
ayo chat read '#room' --follow          # Stream new messages (blocking)
ayo chat read '#room' --json            # Machine readable

# History (alias for read with more context)
ayo chat history '#room'                # Full history
ayo chat history '#room' --json

# Membership
ayo chat who '#room'                    # List room members
ayo chat invite '#room' @agent          # Invite agent to room
ayo chat join '#room'                   # Join room (as current agent)
ayo chat leave '#room'                  # Leave room
```

## Implementation

1. Add cmd/ayo/chat.go with subcommands
2. Add daemon RPC methods:
   - ChatRooms, ChatCreate, ChatSend, ChatRead, ChatWho, ChatInvite, ChatJoin
3. Daemon methods call MatrixBroker
4. Handle agent identity (who is running this command?)
   - Use AYO_AGENT_HANDLE env var if set
   - Otherwise use @ayo

## Files to modify/create

- cmd/ayo/chat.go (new)
- internal/daemon/protocol.go (add RPC types)
- internal/daemon/server.go (add RPC handlers)
- internal/daemon/matrix_broker.go (expose methods)

## Acceptance Criteria

- All commands work as documented
- Works from inside sandboxes
- --json output is parseable
- --follow blocks and streams
- Agent identity correctly determined

