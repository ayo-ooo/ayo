---
id: ayo-ql70
status: closed
deps: [ayo-dtvx]
links: []
created: 2026-02-06T22:16:18Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [testing, manual-test, cleanup]
---
# Test Section 12: Full System Teardown

Complete system teardown and verification of clean state.

## Scope
Complete teardown:
- Stop all services
- Remove all sandboxes
- Verify clean state
- Remove test artifacts

## Test Steps

### 12.1 Stop All Services
```bash
ayo daemon stop 2>/dev/null || true
```
Expected: Daemon stopped

### 12.2 Remove All Sandboxes
```bash
ayo sandbox prune --all --force
ayo sandbox list
```
Expected: All sandboxes removed, list empty

### 12.3 Verify Clean State
```bash
ayo sandbox list
ayo status
ayo triggers list 2>/dev/null || true
```
Expected: No sandboxes, daemon not running, no triggers

### 12.4 Remove Test Artifacts
```bash
rm -rf /tmp/ayo-mount-test /tmp/ayo-mount-readonly /tmp/ayo-watch-test
rm -f /tmp/ayo-test-*
```
Expected: All temp files removed

## Exit Criteria
System returned to clean pre-test state


## Notes

**2026-02-06T22:30:30Z**

COMPLETED:
- 12.1 Stop all services: PASS (used pkill since ayo daemon stop doesn't work)
- 12.2 Remove all sandboxes: PASS
- 12.3 Verify clean state: PASS - no sandboxes, daemon not running
- 12.4 Remove test artifacts: PASS
