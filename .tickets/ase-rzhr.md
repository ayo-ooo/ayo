---
id: ase-rzhr
status: closed
deps: [ase-95o4]
links: []
created: 2026-02-06T04:08:41Z
type: epic
priority: 1
assignee: Alex Cabrera
parent: ase-95o4
---
# Daemon Enhancement

Enhance the ayo daemon to manage sandbox lifecycle, trigger execution, and autonomous agent sessions.

## Design

## Responsibilities
1. Sandbox lifecycle (start, health, restart)
2. Trigger engine (cron, file watch, webhooks, IRC)
3. Agent session management (spawn, track, cleanup)
4. IRC bridge (monitor logs, relay notifications)
5. Git sync (periodic background sync)
6. Memory service (embeddings, search for sandbox tools)

## New Daemon Features
- Trigger registration and execution
- Agent wake/sleep commands
- IRC log monitoring
- Webhook server for external triggers

## Acceptance Criteria

- Daemon starts sandbox automatically
- Triggers fire and spawn agent sessions
- Agents can run autonomously without user
- IRC activity monitored for escalation
- Git sync runs in background

