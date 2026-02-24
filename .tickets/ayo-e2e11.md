---
id: ayo-e2e11
status: open
deps: [ayo-e2e10]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 9 - Full Orchestration Scenarios

## Summary

Write Section 9 of the E2E Manual Testing Guide covering end-to-end orchestration scenarios that test the complete system working together.

## Content Requirements

### Scenario 1: Complex Multi-Step Task
```bash
# Give @ayo a complex task that requires:
# - File creation
# - Code generation
# - Verification

./ayo @ayo "Create a simple calculator in Python with add, subtract, multiply, divide functions. Include tests. Put everything in workspace."

# Verification:
ls ~/.local/share/ayo/workspace/
# Expected: calculator.py, test_calculator.py (or similar)

# Run the tests (in sandbox)
./ayo sandbox exec <sandbox-id> "cd /workspace && python -m pytest"
```

### Scenario 2: Feature Build with Squad
```bash
# Create a feature squad
./ayo squad create feature-build \
  -a @ayo,@tester \
  --output /tmp/feature-output

# Configure SQUAD.md for feature development
cat > ~/.local/share/ayo/sandboxes/squads/feature-build/SQUAD.md << 'EOF'
---
lead: "@ayo"
planners:
  long_term: "ayo-tickets"
agents:
  - "@ayo"
  - "@tester"
---
# Feature Build Squad

## Mission
Build and test a REST API endpoint.

## Agents
### @ayo
**Role**: Developer
### @tester  
**Role**: QA

## Coordination
1. @ayo implements feature
2. @tester writes and runs tests
EOF

# Start squad
./ayo squad start feature-build

# Dispatch feature request
./ayo "#feature-build" "Build a /health endpoint that returns {status: 'ok'}"

# Verify tickets created
./ayo squad ticket feature-build list

# Verify output
cat ~/.local/share/ayo/sandboxes/squads/feature-build/workspace/*.py

# Cleanup
./ayo squad destroy feature-build --delete-data
```

### Scenario 3: Memory-Driven Task
```bash
# Store project context
./ayo memory store "This project uses FastAPI framework"
./ayo memory store "All endpoints must have OpenAPI docs"
./ayo memory store "Use pydantic for validation"

# Now ask for implementation (should use context)
./ayo "Create a user registration endpoint"

# Verify: Response should mention FastAPI, pydantic, OpenAPI
# Agent should have retrieved relevant memories
```

### Scenario 4: Ticket-Driven Development
```bash
# Create project tickets
DESIGN_ID=$(./ayo ticket create "Design user schema" --assignee @ayo --priority high --json | jq -r '.id')
./ayo ticket create "Implement user endpoints" --assignee @ayo --depends-on $DESIGN_ID
./ayo ticket create "Write user tests" --assignee @tester --depends-on $DESIGN_ID

# Check ready queue
./ayo ticket ready
# Expected: Only design ticket

# Complete design
./ayo ticket start $DESIGN_ID
./ayo @ayo "Design a simple user schema with id, email, name fields"
./ayo ticket close $DESIGN_ID

# Check ready queue now
./ayo ticket ready
# Expected: Both implementation and tests are now ready

# Work through remaining tickets
./ayo ticket ready --json | jq -r '.[0].id' | xargs ./ayo ticket start
# ... complete work ...
```

### Scenario 5: Autonomous Workflow
```bash
# Give @ayo a high-level goal and let it orchestrate
./ayo @ayo "Create a squad called 'api-team', design a simple REST API with 3 endpoints, implement them, and write tests. Use tickets to track progress."

# Monitor progress
./ayo squad list
./ayo squad ticket api-team list

# Verify completion
ls ~/.local/share/ayo/sandboxes/squads/api-team/workspace/

# Review what was built
cat ~/.local/share/ayo/sandboxes/squads/api-team/workspace/README.md
```

### Scenario 6: File Sharing Integration
```bash
# Share existing project
mkdir -p /tmp/existing-project
echo "# Existing Project" > /tmp/existing-project/README.md
echo "TODO: add features" > /tmp/existing-project/TODO.md

./ayo share /tmp/existing-project --as project

# Verify share
./ayo share list

# Use in conversation
./ayo "Read the TODO.md in /workspace/project and implement the first TODO item"

# Verify changes
cat /tmp/existing-project/*.py  # or whatever was created

# Cleanup
./ayo share rm project
```

### Verification Criteria
- [ ] Multi-step tasks complete successfully
- [ ] Squad-based development works end-to-end
- [ ] Memory context influences responses
- [ ] Ticket workflow enforces dependencies
- [ ] Autonomous orchestration produces results
- [ ] File sharing enables real project work

## Acceptance Criteria

- [ ] Section written in guide
- [ ] All 6 scenarios documented with commands
- [ ] Verification steps for each scenario
- [ ] Cleanup instructions included
- [ ] Realistic use cases demonstrated
