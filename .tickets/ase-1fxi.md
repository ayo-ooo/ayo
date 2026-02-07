---
id: ase-1fxi
status: closed
deps: []
links: []
created: 2026-02-06T04:11:49Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-py58
---
# Default sandbox-enabled for all agents

Change agent configuration to enable sandbox by default. All agents should execute in the sandbox unless explicitly disabled.

## Design

## Current State
Agent SandboxConfig has 'Enabled' field that defaults to false.

## Changes
1. In agent/agent.go, default SandboxConfig.Enabled to true
2. Update agent loading to apply sandbox defaults
3. Add 'sandbox.enabled: false' to explicitly disable

## Default SandboxConfig
Enabled: true
User: agent handle (e.g., 'ayo' for @ayo)
PersistHome: true
Network: enabled by default

## @ayo Built-in
Update internal/builtin/agents/@ayo/config.json to have sandbox section (even though default, explicit is clearer).

## Backward Compatibility
Existing agent configs without sandbox section should work - they get the new defaults.
Agents with 'sandbox.enabled: false' continue to run on host.

## Acceptance Criteria

- New agents have sandbox enabled by default
- @ayo runs in sandbox
- sandbox.enabled: false disables sandbox
- Tests updated for new defaults

