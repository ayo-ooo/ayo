---
id: ayo-hitv
status: open
deps: [ayo-htol, ayo-htui, ayo-hcht, ayo-heml, ayo-hval, ayo-htim, ayo-hper]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 3
assignee: Alex Cabrera
parent: ayo-hitl
tags: [human-in-the-loop, verification, e2e]
---
# Task: Human-in-the-Loop E2E Verification

## Summary

End-to-end verification that the human-in-the-loop system works correctly across all interfaces and scenarios.

## Test Scenarios

### 1. CLI Form Test
```bash
# Agent requests approval
ayo run @test-agent "Ask me to approve a test request"

# Verify:
# - Form renders with huh
# - All field types work
# - Validation triggers re-prompt
# - Submit returns values to agent
```

### 2. Interactive Chat Test
```bash
# Start interactive mode
ayo chat @test-agent

# Trigger input request
> Request my approval for something

# Verify:
# - Conversational flow works
# - Numbered options presented
# - Invalid input re-prompts
# - Completion acknowledged
```

### 3. Telegram Test
```bash
# Configure Telegram trigger
ayo trigger add telegram-test --type telegram --agent @test-agent

# Send message to bot
# Verify conversational input works
```

### 4. Email Test (Manual)
```
# Configure IMAP/SMTP
# Agent sends input request email
# Reply with keyword
# Verify response parsed correctly
```

### 5. Timeout Test
```bash
# Configure short timeout
ayo run @test-agent "Ask for input with 5s timeout"

# Don't respond
# Verify fallback action triggers
```

### 6. Persona Test
```bash
# Configure persona
# Send request to third party
# Verify no AI disclosure
# Send request to owner
# Verify disclosure (if configured)
```

## Acceptance Criteria

- [ ] CLI forms render and submit correctly
- [ ] Chat input works in interactive mode
- [ ] Chat input works via Telegram trigger
- [ ] Email input sends and receives
- [ ] Timeout fallback works
- [ ] Persona management correct
- [ ] Validation re-prompts work
- [ ] All field types functional
- [ ] Audit logging captures all interactions

## Verification Script

```bash
#!/bin/bash
# Run full HITL verification suite

echo "Testing CLI forms..."
ayo run @test-agent "test_cli_form" --timeout 30s

echo "Testing chat input..."
ayo chat @test-agent --test-input "test_chat_form"

echo "Testing timeout..."
ayo run @test-agent "test_timeout" --timeout 5s

echo "All tests complete"
```
