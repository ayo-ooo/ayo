---
id: ase-py58
status: closed
deps: [ase-95o4, ase-ka3q]
links: []
created: 2026-02-06T04:08:49Z
type: epic
priority: 1
assignee: Alex Cabrera
parent: ase-95o4
---
# Sandbox-by-Default Execution

Modify the run package and agent execution to use sandbox by default. All bash tool calls execute in the sandbox unless explicitly disabled.

## Design

## Changes Required
1. Agent config defaults sandbox.enabled=true
2. Run package acquires sandbox before execution
3. Bash tool routes through sandbox executor
4. Session workspace created in /workspaces/{session}/
5. Host-side tools (memory, agent_call) intercepted before dispatch

## Tool Routing
- bash → sandbox
- memory → host (intercept)
- agent_call → host (intercept)
- file_request → host+sandbox (bridge)
- publish → host+sandbox (bridge)

## Acceptance Criteria

- Agents execute in sandbox by default
- Sandbox can be disabled with sandbox.enabled=false
- Memory/agent_call tools work from sandbox context
- File transfer tools bridge host and sandbox

