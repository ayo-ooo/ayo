---
id: ase-mwdy
status: closed
deps: [ase-w2n6]
links: []
created: 2026-02-09T03:05:02Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-qnjh
---
# Implement Matrix message broker in daemon

Implement a message broker in the daemon that maintains a single Matrix sync connection and routes messages to agents.

## Background

Agents don't maintain their own Matrix connections. The daemon:
1. Maintains ONE sync connection to Conduit (via mautrix-go)
2. Routes messages to agents based on room membership
3. Spawns agent processes when they receive messages
4. Collects responses

This model means agents are stateless functions, not persistent processes.

## Architecture

```
Daemon
  ├── MatrixBroker
  │   ├── mautrix.Client (single sync connection)
  │   ├── RoomMembership map[RoomID][]AgentHandle
  │   └── MessageQueue per agent
  │
  └── On message received:
      1. Look up room → agent mapping
      2. For each agent in room:
         a. If message is for them (mentions or broadcast)
         b. Invoke agent with message context
         c. Agent reads history via CLI
         d. Agent responds via CLI
         e. Broker sends response to room
```

## Implementation

1. Add mautrix-go dependency
2. Create MatrixBroker struct:
   - Connect(ctx) - establish sync connection
   - CreateRoom(name) - create session room
   - InviteAgent(room, agent) - add agent to room
   - SendMessage(room, content) - send as @ayo or specified user
   - OnMessage callback registration
3. Integrate with agent invocation:
   - Set AYO_SESSION_ROOM env var
   - Set AYO_MATRIX_MESSAGE env var (the triggering message)
4. Handle agent registration:
   - Each agent gets a Matrix user: @{handle}:ayo.local
   - Auto-register on first use

## Files to modify/create

- go.mod (add mautrix-go)
- internal/daemon/matrix_broker.go (new)
- internal/daemon/server.go (initialize broker)
- internal/daemon/agent_invoke.go (pass Matrix context)

## Acceptance Criteria

- Single sync connection works
- Messages routed to correct agents
- Agents can be invoked with Matrix context
- Room creation and membership management works
- Agent registration is automatic

