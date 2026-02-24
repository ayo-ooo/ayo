---
id: ayo-u43p
status: open
deps: [ayo-dicu, ayo-c5mt, ayo-evik, ayo-vclt, ayo-66df]
links: []
created: 2026-02-24T01:02:41Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-whmn
tags: [verification, e2e]
---
# Phase 2 E2E verification

End-to-end verification that file system model works correctly.

## Prerequisites

All Phase 2 tickets complete:
- file_request tool (ayo-dicu)
- Approval UI (ayo-c5mt)
- --no-jodas flag (ayo-evik)
- Audit logging (ayo-vclt)
- /output sync (ayo-66df)

## Verification Checklist

### file_request Tool
- [ ] Agent can call file_request tool
- [ ] Request appears in terminal UI
- [ ] [Y]es approves and writes file
- [ ] [N]o denies and returns error to agent
- [ ] [D]iff shows unified diff
- [ ] [A]lways caches approval for session

### --no-jodas Mode
- [ ] `ayo --no-jodas "edit file"` auto-approves
- [ ] Global config `permissions.no_jodas` works
- [ ] Per-agent `permissions.auto_approve` works
- [ ] Precedence: session cache > CLI > agent > global

### Audit Logging
- [ ] `~/.local/share/ayo/audit.log` exists
- [ ] Each modification logged with timestamp, agent, path
- [ ] Approval method recorded (user, no_jodas, etc)
- [ ] `ayo audit list` shows entries

### /output Safe Write Zone
- [ ] `/output/` exists in sandbox
- [ ] Agent can write to `/output/` without approval
- [ ] Files sync to `~/.local/share/ayo/output/`
- [ ] $OUTPUT env var set correctly

### Host Mount
- [ ] `/mnt/{username}` readable in sandbox
- [ ] Cannot write to `/mnt/{username}` directly
- [ ] file_request required for host writes

## Acceptance Criteria

All checkboxes verified. File system model is intuitive and secure.
