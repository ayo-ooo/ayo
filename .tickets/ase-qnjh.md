---
id: ase-qnjh
status: open
deps: []
links: []
created: 2026-02-09T03:03:16Z
type: epic
priority: 2
assignee: Alex Cabrera
parent: ase-zlew
---
# Matrix Integration

Integrate Matrix (via Conduit homeserver) for inter-agent communication. The daemon runs a lightweight Matrix server, routes messages to agents, and provides CLI access for agents in sandboxes.

## Acceptance Criteria

- Conduit homeserver embedded in ayo service
- Daemon maintains single sync connection
- Agents invoked on message receipt
- ayo chat CLI for agent communication
- Human can optionally connect via Element
- Session rooms created per orchestration

