---
id: ayo-e2e13
status: open
deps: [ayo-e2e12]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 11 - Cleanup & Final Verification

## Summary

Write Section 11 of the E2E Manual Testing Guide covering cleanup procedures and final system verification.

## Content Requirements

### Test Artifacts Cleanup
```bash
# List all test squads
./ayo squad list

# Destroy each test squad
./ayo squad destroy e2e-squad --delete-data
./ayo squad destroy feature-build --delete-data
./ayo squad destroy api-team --delete-data
# etc.

# Verify
./ayo squad list
# Expected: No test squads remaining
```

### Remove Test Agents
```bash
# List agents
./ayo agents list

# Remove test agents (keep defaults)
./ayo agents rm @tester
./ayo agents rm @tester-verbose
# etc.

# Verify only defaults remain
./ayo agents list
# Expected: @ayo and other built-in agents only
```

### Clear Test Tickets
```bash
# List tickets
./ayo ticket list

# Remove test tickets or mark as closed
./ayo ticket close <id>
# or
./ayo ticket rm <id>

# Verify
./ayo ticket list
# Expected: No test tickets
```

### Clear Test Memories
```bash
./ayo memory list

# Remove test memories
./ayo memory rm <id>
# or clear all (with confirmation)
./ayo memory clear --confirm

# Verify
./ayo memory list
# Expected: Empty or only essential memories
```

### Remove Test Triggers
```bash
./ayo triggers list

# Remove any test triggers
./ayo triggers rm <id>

# Verify
./ayo triggers list
# Expected: No test triggers
```

### Clean Temporary Files
```bash
# Remove any temp files created during testing
rm -rf /tmp/e2e-*
rm -rf /tmp/existing-project
rm -rf /tmp/feature-output
rm -rf /tmp/watched
rm -f /tmp/test.txt
```

### Stop Daemon
```bash
./ayo sandbox service stop

# Verify stopped
./ayo sandbox service status
# Expected: Not running
```

### Final Health Check
```bash
# Start daemon fresh
./ayo sandbox service start

# Run full doctor check
./ayo doctor -v

# Expected:
# ✓ Configuration valid
# ✓ Daemon running
# ✓ Container runtime available
# ✓ Provider configured
# ✓ Default agents installed
# ✓ No orphaned resources
```

### Full Clean State (Optional)
```bash
# If desired, return to completely clean state
./ayo sandbox service stop

# Run clean state script from Section 0
./clean-state.sh

# Verify clean
ls ~/.local/share/ayo/
# Expected: Minimal or empty

# Reinstall if continuing
./ayo setup
./ayo sandbox service start
```

### Final Verification Checklist

| Item | Status | Notes |
|------|--------|-------|
| All test squads destroyed | ☐ | |
| All test agents removed | ☐ | |
| All test tickets cleared | ☐ | |
| All test memories cleared | ☐ | |
| All test triggers removed | ☐ | |
| Temp files cleaned | ☐ | |
| Daemon running cleanly | ☐ | |
| Doctor passes | ☐ | |
| No orphaned sandboxes | ☐ | |
| No orphaned sessions | ☐ | |

### E2E Testing Complete

**If all sections passed:** System is verified working end-to-end.

**If any section failed:**
1. Note which section failed
2. Run clean-state.sh
3. Re-run `ayo setup`
4. Restart from Section 1
5. Debug the failing section

## Acceptance Criteria

- [ ] Section written in guide
- [ ] Complete cleanup procedure documented
- [ ] Final verification checklist included
- [ ] Pass/fail determination criteria clear
- [ ] Recovery instructions for failures
