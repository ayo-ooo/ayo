---
id: ase-fw7m
status: closed
deps: [ase-ka3q]
links: []
created: 2026-02-06T04:11:01Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ase-rzhr
---
# Add sandbox lifecycle management to daemon

The daemon should automatically start the persistent sandbox on startup and manage its lifecycle.

## Design

## Current State
Daemon has a sandbox pool but doesn't auto-start a persistent sandbox.

## New Behavior
1. On daemon start:
   - Check if persistent sandbox exists
   - If not, create it with standard Alpine config
   - If stopped, restart it
   - Ensure ngircd and other services are running

2. Health monitoring:
   - Periodic health check (every 30s)
   - Restart if unresponsive
   - Log issues

3. On daemon stop:
   - Keep sandbox running (persistent)
   - OR stop sandbox (configurable)

## Persistent Sandbox Config
Name: 'ayo-sandbox' (fixed name)
Image: Alpine with ngircd
Mounts:
  - ~/.local/share/ayo/sandbox/homes:/home
  - ~/.local/share/ayo/sandbox/shared:/shared
  - ~/.local/share/ayo/sandbox/workspaces:/workspaces
  - ~/.local/share/ayo/sandbox/irc-logs:/var/log/irc

## Implementation
1. Add ensurePersistentSandbox() to Server
2. Call on Start()
3. Add health check goroutine
4. Store sandbox ID in Server struct

## Acceptance Criteria

- Sandbox starts automatically with daemon
- Sandbox survives daemon restarts
- Health check restarts unhealthy sandbox
- Standard directories mounted from host

