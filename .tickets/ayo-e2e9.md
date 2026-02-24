---
id: ayo-e2e9
status: closed
deps: [ayo-e2e8]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 7 - Triggers & Scheduling

## Summary

Write Section 7 of the E2E Manual Testing Guide covering trigger creation, scheduling, and automation.

## Content Requirements

### Cron Schedule Triggers
```bash
# Create schedule trigger (every minute for testing)
./ayo triggers schedule @tester "*/1 * * * *" \
  --prompt "Log a test message: 'Trigger fired at $(date)'"

# Expected: Trigger ID returned

# List triggers
./ayo triggers list
# Expected: Shows the schedule trigger
```

### Watch Triggers (File System)
```bash
# Create watch directory
mkdir -p /tmp/e2e-watch

# Create watch trigger
./ayo triggers watch /tmp/e2e-watch @tester \
  --prompt "A file was changed in the watched directory"

# Verify trigger
./ayo triggers list
# Expected: Shows watch trigger
```

### Trigger Execution Verification
```bash
# For schedule trigger: wait 1-2 minutes
# Then check session history or logs

# For watch trigger: create a file
touch /tmp/e2e-watch/test-file.txt

# Verify trigger fired (check recent sessions)
./ayo session list
# Expected: New session from trigger execution
```

### Trigger Details
```bash
# Show trigger details
./ayo triggers show <trigger-id>

# Expected:
# - Type (schedule/watch)
# - Agent
# - Prompt
# - Schedule/Path
# - Last fired
# - Next fire (for schedule)
```

### Trigger Management
```bash
# Disable trigger (pause)
./ayo triggers disable <trigger-id>

# Verify disabled
./ayo triggers list
# Expected: Shows disabled status

# Enable trigger
./ayo triggers enable <trigger-id>

# Verify enabled
./ayo triggers list
```

### Trigger Removal
```bash
# Remove triggers
./ayo triggers rm <schedule-trigger-id>
./ayo triggers rm <watch-trigger-id>

# Verify removal
./ayo triggers list
# Expected: Triggers no longer present

# Cleanup watch directory
rm -rf /tmp/e2e-watch
```

### Advanced: Event Triggers (if applicable)
```bash
# Trigger on ticket events
./ayo triggers on ticket:created @notifier \
  --prompt "New ticket created: {{ticket.title}}"

# Trigger on squad events
./ayo triggers on squad:started @monitor \
  --prompt "Squad {{squad.name}} started"
```

### Verification Criteria
- [ ] Schedule trigger creation works
- [ ] Watch trigger creation works
- [ ] Triggers fire correctly (verify via sessions)
- [ ] Trigger listing shows all triggers
- [ ] Trigger disable/enable works
- [ ] Trigger removal works

## Acceptance Criteria

- [ ] Section written in guide
- [ ] Both trigger types documented
- [ ] Trigger execution verified
- [ ] Complete lifecycle documented
- [ ] Cleanup instructions included
