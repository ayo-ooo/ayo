---
id: ayo-ut3g
status: closed
deps: []
links: []
created: 2026-02-06T22:20:39Z
type: bug
priority: 2
assignee: Alex Cabrera
tags: [sandbox, daemon]
---
# Sandbox pool count inconsistent with actual containers

## Issue
`ayo status` shows Sandbox Pool with Total:1, Idle:1 but:
- `ayo sandbox list` shows 'No active sandboxes'
- `container list --all` shows no containers

## Reproduction
1. Start daemon: `ayo daemon start`
2. Sandbox is created
3. Stop and prune the sandbox
4. `ayo status` still shows Total:1, Idle:1
5. `ayo sandbox list` shows no sandboxes

## Expected
Pool count should match actual container count, or pool should auto-recreate containers.

## Actual
Pool reports containers exist that don't actually exist in the runtime.

