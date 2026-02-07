---
id: ase-dp2s
status: closed
deps: [ase-t4cr]
links: []
created: 2026-02-06T04:14:36Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-x0vq
---
# Add agent status/wake/sleep commands

Add CLI commands for managing agent sessions: status, wake, sleep.

## Design

## Commands
ayo agents status              # List active agent sessions
ayo agents wake @ayo           # Start session for agent
ayo agents sleep @crush        # Stop agent session

## agents status
Output:
  Agent    Status      Started      Last Active
  @ayo     running     10m ago      2s ago
  @crush   idle        1h ago       30m ago
  @research sleeping   -            1d ago

## agents wake
- Check if agent has active session
- If not, start new session
- If daemon not running, start it

## agents sleep
- Find agent's active session
- Gracefully stop it
- Agent can be woken again

## Implementation
1. Add commands to cmd/ayo/agents.go
2. Implement daemon methods for session management
3. CLI calls daemon via client

## Daemon Integration
Uses daemon's session manager:
- MethodSessionList
- MethodSessionStart (wake)
- MethodSessionStop (sleep)

## Flags
--json: JSON output
--quiet: Minimal output

## Acceptance Criteria

- Status shows agent sessions
- Wake starts agent session
- Sleep stops agent session
- Works via daemon

