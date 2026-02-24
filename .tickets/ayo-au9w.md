---
id: ayo-au9w
status: closed
deps: [ayo-enaj, ayo-kkxg, ayo-ao4q, ayo-1xg8, ayo-c8px]
links: []
created: 2026-02-24T01:01:35Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [verification, e2e]
---
# Phase 1 E2E verification

End-to-end verification that Phase 1 is complete and no regressions occurred.

## Prerequisites

All Phase 1 tickets must be complete:
- Code removal (ayo-ydub, ayo-8nn8, ayo-rdao, ayo-tha0, ayo-1ryh, ayo-ieiy, ayo-c9zl)
- Code cleanup (ayo-fwye, ayo-qbsu, ayo-enaj)
- Sandbox bootstrap (ayo-kkxg, ayo-ao4q, ayo-1xg8, ayo-c8px)

## Verification Checklist

### Build & Test
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes (all tests green)
- [ ] `golangci-lint run` passes

### Removed Commands
- [ ] `ayo serve` returns "unknown command"
- [ ] `ayo chat` returns "unknown command"
- [ ] `ayo flow run` returns "unknown command"

### Daemon
- [ ] `ayo daemon start` starts without error
- [ ] `ayo daemon status` shows running
- [ ] Daemon logs show no errors

### Sandbox with ayod
- [ ] `ayo sandbox create test-sb` creates sandbox
- [ ] ayod running as PID 1 inside sandbox
- [ ] `ayo sandbox exec test-sb -- whoami` returns "ayo"
- [ ] User creation works: new agent gets Unix user
- [ ] `/workspace` exists with correct permissions

### Host Mount
- [ ] `/mnt/{username}` mounted inside sandbox
- [ ] Read-only: cannot write to `/mnt/{username}`
- [ ] Can read files from host home

### Code Metrics
- [ ] Line count reduced by ~5000 lines
- [ ] go.mod has fewer dependencies
- [ ] No references to removed packages

## Acceptance Criteria

All checkboxes above must be verified. Document any issues found and create tickets for fixes.
