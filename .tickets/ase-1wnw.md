---
id: ase-1wnw
status: closed
deps: [ase-1fxi]
links: []
created: 2026-02-06T04:12:00Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-py58
---
# Route bash tool through sandbox executor

Modify the bash tool execution to always use the sandbox when enabled. The run package should acquire a sandbox and execute commands there.

## Design

## Current State
run/fantasy_tools.go creates bash tool that executes locally or via sandbox provider if configured.

## Changes
1. When agent has sandbox enabled, Runner MUST have sandbox provider
2. Before first tool call, ensure sandbox is acquired
3. Ensure agent user exists in sandbox
4. Route all bash calls through sandbox.Exec()

## Session-Sandbox Binding
Each session should be bound to a sandbox.
Store sandbox ID in session context.

## User Execution
Commands execute as the agent's user:
- @ayo → runs as 'ayo' user
- @crush → runs as 'crush' user

## Working Directory
Default to /workspaces/{session}/ for session-scoped work.
If project mounted, could be /mnt/host/project/

## Implementation
1. Add ensureSandbox() to Runner
2. Call before any tool execution
3. Pass sandbox ID to BashExecutor
4. Update BashExecutor to use sandbox.Exec with user

## Acceptance Criteria

- Bash commands execute in sandbox
- Commands run as agent's user
- Working directory is session workspace
- Sandbox acquired on first tool call

