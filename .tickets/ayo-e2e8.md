---
id: ayo-e2e8
status: closed
deps: [ayo-e2e7]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 6 - Squads & Coordination

## Summary

Write Section 6 of the E2E Manual Testing Guide covering squad creation, SQUAD.md constitution, and multi-agent coordination.

## Content Requirements

### Create Squad
```bash
./ayo squad create e2e-squad \
  --description "E2E testing squad"

# Expected: Squad created with default SQUAD.md
```

### Verify Squad Structure
```bash
# Check squad directory
ls -la ~/.local/share/ayo/sandboxes/squads/e2e-squad/

# Expected structure:
# ├── SQUAD.md          ← Constitution
# ├── .tickets/         ← Squad-scoped tickets
# ├── .context/         ← Persistent state
# ├── workspace/        ← Shared code
# └── agent-homes/      ← Per-agent directories
```

### Edit SQUAD.md Constitution
```bash
# View default SQUAD.md
cat ~/.local/share/ayo/sandboxes/squads/e2e-squad/SQUAD.md

# Edit constitution
cat > ~/.local/share/ayo/sandboxes/squads/e2e-squad/SQUAD.md << 'EOF'
---
lead: "@ayo"
planners:
  long_term: "ayo-tickets"
agents:
  - "@ayo"
  - "@tester"
---
# Squad: e2e-squad

## Mission

Test the complete ayo system including squad coordination, ticket management, and workspace operations.

## Context

- This is an E2E test squad
- All operations should be logged for verification
- Use workspace for any file outputs

## Agents

### @ayo
**Role**: Squad lead and coordinator
**Responsibilities**:
- Create tickets for work items
- Delegate to @tester
- Verify completion

### @tester
**Role**: Testing agent
**Responsibilities**:
- Execute test tasks
- Report results
- Create test files in workspace

## Coordination

1. @ayo creates tickets for test tasks
2. @tester picks up ready tickets
3. @tester completes work and closes tickets
4. @ayo verifies results

## Guidelines

- All output files go in workspace/
- Use descriptive ticket titles
- Mark tickets with [TEST] prefix
EOF

# Verify changes
cat ~/.local/share/ayo/sandboxes/squads/e2e-squad/SQUAD.md
```

### Add/Remove Agents
```bash
# Add agent
./ayo squad add-agent e2e-squad @tester

# Verify
./ayo squad show e2e-squad
# Expected: Shows @tester in agent list

# Remove agent
./ayo squad remove-agent e2e-squad @tester

# Add back for testing
./ayo squad add-agent e2e-squad @tester
```

### Start/Stop Squad
```bash
# Start squad sandbox
./ayo squad start e2e-squad

# Verify running
./ayo squad status e2e-squad
# Expected: Status = running

# Stop squad
./ayo squad stop e2e-squad

# Restart for further testing
./ayo squad start e2e-squad
```

### Squad Ticket Management
```bash
# Create ticket in squad
./ayo squad ticket e2e-squad create "[TEST] Create hello.txt" \
  --assignee @tester

# List squad tickets
./ayo squad ticket e2e-squad list

# Verify ticket
./ayo squad ticket e2e-squad show <id>
```

### Dispatch to Squad
```bash
# Dispatch task to squad (uses #squad-name syntax)
./ayo "#e2e-squad" "Create a file called hello.txt containing 'Hello from E2E test'"

# Expected: Squad receives task, creates ticket, executes work
```

### Verify Workspace Output
```bash
# Check workspace
ls ~/.local/share/ayo/sandboxes/squads/e2e-squad/workspace/

# Verify file contents
cat ~/.local/share/ayo/sandboxes/squads/e2e-squad/workspace/hello.txt
# Expected: "Hello from E2E test"
```

### Multi-Agent Coordination Test
```bash
# Create multi-step task
./ayo "#e2e-squad" "Have @tester create a test plan document and then @ayo verify it exists"

# Verify coordination
./ayo squad ticket e2e-squad list
# Expected: Multiple tickets showing coordination
```

### Squad Destruction
```bash
# Stop squad first
./ayo squad stop e2e-squad

# Destroy (keeps data)
./ayo squad destroy e2e-squad

# Or destroy with data deletion
./ayo squad destroy e2e-squad --delete-data

# Verify removal
./ayo squad list
# Expected: e2e-squad not present
```

### Verification Criteria
- [ ] Squad creation works
- [ ] SQUAD.md constitution editable
- [ ] Agent add/remove works
- [ ] Squad start/stop works
- [ ] Squad tickets work
- [ ] Dispatch to squad works
- [ ] Workspace output verified
- [ ] Multi-agent coordination works
- [ ] Squad destruction works

## Acceptance Criteria

- [ ] Section written in guide
- [ ] Complete squad lifecycle documented
- [ ] SQUAD.md editing documented
- [ ] Dispatch syntax (#squad-name) tested
- [ ] Workspace verification included
- [ ] Multi-agent scenario tested
