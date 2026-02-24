---
id: ayo-e2e4
status: closed
deps: [ayo-e2e3]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 2 - Agent Management

## Summary

Write Section 2 of the E2E Manual Testing Guide covering agent creation, configuration, and management.

## Content Requirements

### List Default Agents
```bash
./ayo agents list

# Expected: Shows @ayo (meta-agent) and other defaults
```

### View Agent Details
```bash
./ayo agents show @ayo

# Expected output:
# - Agent handle
# - Description
# - System prompt (or prompt path)
# - Provider/model configuration
# - Capabilities
```

### Create Custom Agent
```bash
./ayo agents create @tester \
  --description "E2E test agent" \
  --prompt "You are a testing assistant. Help verify system functionality." \
  --model claude-sonnet-4-20250514

# Verify creation
./ayo agents list
./ayo agents show @tester
```

### Configure Agent (Edit JSON)
```bash
# Find agent config
ls ~/.local/share/ayo/agents/

# Edit agent
cat ~/.local/share/ayo/agents/tester.json

# Modify (example: add tools)
# ... edit file ...

# Verify changes
./ayo agents show @tester
```

### Test Agent
```bash
# Quick test
./ayo @tester "Hello, what is your purpose?"

# Expected: Agent responds with testing-related description
```

### Agent Inheritance
```bash
# Create agent that inherits from another
./ayo agents create @tester-verbose \
  --description "Verbose tester" \
  --inherits @tester

# Verify inheritance
./ayo agents show @tester-verbose
```

### Remove Agent
```bash
./ayo agents rm @tester-verbose

# Verify removal
./ayo agents list
# @tester-verbose should not appear
```

### Verification Criteria
- [ ] Default agents present after setup
- [ ] Can view agent details
- [ ] Custom agent creation works
- [ ] Agent responds to prompts
- [ ] Agent inheritance works
- [ ] Agent deletion works

## Acceptance Criteria

- [ ] Section written in guide
- [ ] All CRUD operations documented
- [ ] Agent inheritance tested
- [ ] Cleanup instructions included
