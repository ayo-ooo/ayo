---
id: ayo-e2e5
status: open
deps: [ayo-e2e4]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 3 - Sessions & Chat

## Summary

Write Section 3 of the E2E Manual Testing Guide covering interactive chat, sessions, and conversation management.

## Content Requirements

### Interactive Chat
```bash
# Start interactive session
./ayo

# Expected: TUI opens, can type messages
# Type: "What is 2+2?"
# Expected: Agent responds with "4"

# Exit with /exit or Ctrl+D
```

### Single Prompt (Quoted)
```bash
# Quoted prompt executes without entering TUI
./ayo "What is the capital of France?"

# Expected: Paris (and returns to shell)
```

### Prompt Detection Rules
```bash
# Single lowercase word = command (error expected)
./ayo foobar
# Error: unknown command "foobar"

# Quoted = always prompt
./ayo "foobar"
# Agent responds

# Uppercase or multi-word = prompt
./ayo Hello
./ayo Hello world
# Agent responds
```

### Session Continuation
```bash
# First interaction
./ayo "Remember: the secret code is ALPHA-7"

# Continue same session
./ayo -c "What was the secret code?"

# Expected: Agent recalls ALPHA-7
```

### File Attachments
```bash
# Create test file
echo "Hello from E2E test" > /tmp/e2e-test.txt

# Attach file to prompt
./ayo "What does this file contain?" -a /tmp/e2e-test.txt

# Expected: Agent describes file contents
```

### Session Listing
```bash
./ayo session list

# Expected: Shows recent sessions with IDs and timestamps
```

### Session Management
```bash
# View specific session
./ayo session show <session-id>

# Delete session
./ayo session rm <session-id>

# Verify deletion
./ayo session list
```

### Multi-Turn Context
```bash
./ayo @tester
> My name is Alice
> What is my name?
# Expected: "Alice"
> /exit
```

### Verification Criteria
- [ ] Interactive chat works
- [ ] Single quoted prompts work
- [ ] Session continuation works
- [ ] File attachments work
- [ ] Session listing works
- [ ] Multi-turn context maintained

## Acceptance Criteria

- [ ] Section written in guide
- [ ] All chat modes documented
- [ ] Session management tested
- [ ] File attachments tested
- [ ] Context preservation verified
