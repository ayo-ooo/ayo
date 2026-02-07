---
id: ayo-fpxg
status: closed
deps: []
links: []
created: 2026-02-06T22:14:25Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [testing, manual-test]
---
# Test Section 1: Environment Setup

Setup and verify clean test environment for ayo manual testing.

## Scope
- Build fresh ayo binary from source
- Clean all existing sandboxes and daemon state
- Verify Docker is available and running

## Setup Commands
```bash
go build -o ayo ./cmd/ayo
ayo daemon stop 2>/dev/null || true
ayo sandbox prune --all --force 2>/dev/null || true
docker info >/dev/null 2>&1 && echo 'Docker OK'
```

## Verification
- [ ] Binary exists at ./ayo
- [ ] ./ayo --version returns version string
- [ ] ayo sandbox list shows no sandboxes
- [ ] ayo status shows daemon not running
- [ ] Docker is available

## Cleanup
None - this IS the cleanup phase

## Exit Criteria
Clean slate ready for testing


## Notes

**2026-02-06T22:17:06Z**

PASSED: Binary built (v0.3.0), clean state verified, Docker available
