---
id: ayo-e2e7
status: closed
deps: [ayo-e2e6]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 5 - Tickets & Planning

## Summary

Write Section 5 of the E2E Manual Testing Guide covering ticket-based coordination and planning workflows.

## Content Requirements

### Create Tickets
```bash
# Create simple ticket
./ayo ticket create "Implement login endpoint"

# Create with metadata
./ayo ticket create "Design API schema" \
  --assignee @architect \
  --priority high \
  --tags api,design

# Expected: Ticket ID returned (e.g., ayo-a1b2)
```

### List Tickets
```bash
./ayo ticket list

# Expected: Shows tickets with status, assignee, priority
```

### View Ticket Details
```bash
./ayo ticket show ayo-a1b2

# Expected: Full ticket details including:
# - Title
# - Status
# - Assignee
# - Dependencies
# - Created/updated timestamps
```

### Ticket Dependencies
```bash
# Create dependent ticket
./ayo ticket create "Implement login endpoint" \
  --depends-on ayo-a1b2 \
  --assignee @backend

# Verify dependency
./ayo ticket show <new-id>
# Expected: Shows dependency on ayo-a1b2
```

### Ready/Blocked Queries
```bash
# Check what's ready for @architect
./ayo ticket ready --assignee @architect
# Expected: Shows "Design API schema" (no blockers)

# Check what's ready for @backend
./ayo ticket ready --assignee @backend
# Expected: Empty (blocked by schema design)

# Check blocked tickets
./ayo ticket blocked
# Expected: Shows "Implement login endpoint" blocked by ayo-a1b2
```

### Ticket Workflow
```bash
# Start work on ticket
./ayo ticket start ayo-a1b2

# Verify status
./ayo ticket show ayo-a1b2
# Expected: Status = in_progress

# Close ticket
./ayo ticket close ayo-a1b2

# Verify status
./ayo ticket show ayo-a1b2
# Expected: Status = closed

# Check if dependent is now ready
./ayo ticket ready --assignee @backend
# Expected: Shows "Implement login endpoint" (no longer blocked)
```

### Ticket Assignment
```bash
# Reassign ticket
./ayo ticket assign <id> @tester

# Verify assignment
./ayo ticket show <id>
```

### Priority Management
```bash
# Set priority
./ayo ticket priority <id> critical

# List by priority
./ayo ticket list --sort priority
```

### Verification Criteria
- [ ] Ticket creation works
- [ ] Ticket listing works
- [ ] Dependencies correctly block/unblock
- [ ] Ready queue reflects dependencies
- [ ] Status transitions work (start, close)
- [ ] Assignment and priority work

## Acceptance Criteria

- [ ] Section written in guide
- [ ] Complete ticket lifecycle documented
- [ ] Dependency workflow verified
- [ ] Ready/blocked queries tested
- [ ] All metadata operations documented
