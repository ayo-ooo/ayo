---
id: ase-t4cr
status: closed
deps: [ase-fw7m]
links: []
created: 2026-02-06T04:11:22Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-rzhr
---
# Add agent session management to daemon

The daemon should track active agent sessions and provide APIs for managing them.

## Design

## Session Manager
Track all active agent sessions spawned by daemon.

## Data Structures
type AgentSession struct {
    ID          string
    Agent       string
    StartedAt   time.Time
    TriggerID   string // if started by trigger
    Status      string // running, idle, stopped
    LastActive  time.Time
}

## New Daemon Methods
- MethodSessionList: List active sessions
- MethodSessionStart: Start new agent session
- MethodSessionStop: Stop a session
- MethodSessionInject: Inject message into session

## Agent Wake/Sleep
- Wake: Start a session for an agent (if not running)
- Sleep: Gracefully stop an agent's session

## Idle Timeout
Sessions can be configured to auto-stop after idle period.

## Implementation
1. Add sessionManager to Server
2. Track sessions in map
3. Add new daemon methods
4. Integrate with trigger engine (triggers create sessions)

## CLI Integration
- ayo agents status: list active sessions
- ayo agents wake @ayo: start session
- ayo agents sleep @crush: stop session

## Acceptance Criteria

- Active sessions tracked by daemon
- Sessions can be listed/started/stopped
- Wake/sleep commands work
- Idle timeout configured and working

