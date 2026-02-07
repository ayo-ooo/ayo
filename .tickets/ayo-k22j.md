---
id: ayo-k22j
status: closed
deps: [ayo-ivjs]
links: []
created: 2026-02-06T22:14:46Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [testing, manual-test, sandbox]
---
# Test Section 3: Sandbox Lifecycle

Exhaustively test sandbox container lifecycle operations.

## Scope
Test all sandbox lifecycle operations:
- List sandboxes (empty and populated states)
- Create sandbox via agent command
- Inspect sandbox details
- Execute commands in sandbox
- Stop, start, restart sandbox
- Prune stopped and all sandboxes

## Test Steps

### 3.1 List Sandboxes (Empty State)
```bash
ayo sandbox list
```
Expected: 'No active sandboxes' or empty table

### 3.2 Create Sandbox via Agent
```bash
ayo @ayo "run 'echo sandbox-test-marker' and report the output"
```
Expected: Agent responds, output contains 'sandbox-test-marker'

### 3.3 Verify Sandbox Created
```bash
ayo sandbox list
```
Expected: At least one sandbox listed, status 'running'

### 3.4 Inspect Sandbox Details
```bash
ayo sandbox show <sandbox-id>
```
Expected: ID, status, image, created timestamp, mounts shown

### 3.5 Execute Command in Sandbox
```bash
ayo sandbox exec <id> -- uname -a
ayo sandbox exec <id> -- whoami
ayo sandbox exec <id> -- id
```
Expected: Linux kernel info, user identity shown

### 3.6 Stop Sandbox
```bash
ayo sandbox stop <id>
ayo sandbox list
```
Expected: Status changes to 'stopped'

### 3.7 Restart Sandbox
```bash
ayo sandbox start <id>
ayo sandbox list
```
Expected: Status returns to 'running'

### 3.8 Prune Stopped Sandboxes
```bash
ayo sandbox stop <id>
ayo sandbox prune --force
ayo sandbox list
```
Expected: Sandbox removed from list

### 3.9 Prune All Sandboxes
```bash
ayo sandbox prune --all --force
ayo sandbox list
```
Expected: All sandboxes removed

## Analysis Required
- Record sandbox ID format
- Record image used (alpine vs busybox)
- Record mount configuration
- Document user context in sandbox
- Note any errors or warnings

## Cleanup
```bash
ayo sandbox prune --all --force
```

## Exit Criteria
All sandbox lifecycle operations work correctly


## Notes

**2026-02-06T22:20:47Z**

COMPLETED with issues:
- 3.1 List empty: PASS
- 3.2 Create via agent: PASS (ephemeral sandbox created and destroyed)
- 3.3 Verify created: PASS (with daemon running, pool sandbox visible)
- 3.4 Inspect details: PASS (shows ID, status, image, created)
- 3.5 Execute command: PASS (uname, whoami, id work)
- 3.6 Stop sandbox: PASS
- 3.7 Restart sandbox: FAIL - no 'sandbox start' command (ticket ayo-ps71)
- 3.8-3.9 Prune: PASS

Additional issues found:
- Pool count inconsistent with actual containers (ticket ayo-ut3g)
- Image is busybox:stable not alpine as documented
