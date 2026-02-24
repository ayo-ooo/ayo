---
id: ayo-e2e12
status: closed
deps: [ayo-e2e11]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 10 - Error Handling & Edge Cases

## Summary

Write Section 10 of the E2E Manual Testing Guide covering error scenarios, edge cases, and recovery procedures.

## Content Requirements

### Invalid Command Detection
```bash
# Single lowercase word = unknown command
./ayo foobar
# Expected: Error - unknown command "foobar"

# Misspelled command
./ayo agnet list
# Expected: Error - unknown command "agnet" (did you mean "agent"?)

# Correct the command
./ayo agents list
# Expected: Works
```

### Daemon Not Running
```bash
# Stop daemon
./ayo sandbox service stop

# Try sandbox operation
./ayo sandbox list
# Expected: Error - cannot connect to daemon

# Try dispatch
./ayo "#squad-name" "test"
# Expected: Error - daemon not running

# Recovery
./ayo sandbox service start
./ayo sandbox list
# Expected: Works
```

### Unknown Agent
```bash
./ayo @nonexistent "hello"
# Expected: Error - agent "@nonexistent" not found

./ayo agents show @fake
# Expected: Error - agent not found
```

### Invalid Squad
```bash
./ayo "#no-such-squad" "test"
# Expected: Error - squad "no-such-squad" not found

./ayo squad start nosquad
# Expected: Error - squad not found
```

### Invalid Ticket Reference
```bash
./ayo ticket show xyz-9999
# Expected: Error - ticket not found

./ayo ticket close invalid-id
# Expected: Error - ticket not found
```

### Circular Dependencies
```bash
# Create tickets
ID1=$(./ayo ticket create "Task A" --json | jq -r '.id')
ID2=$(./ayo ticket create "Task B" --depends-on $ID1 --json | jq -r '.id')

# Try to create circular dependency
./ayo ticket update $ID1 --depends-on $ID2
# Expected: Error - circular dependency detected
```

### Permission Errors (Sandbox)
```bash
# Try to access protected path
./ayo sandbox exec <id> "cat /etc/shadow"
# Expected: Error or permission denied

# Try to escape sandbox
./ayo sandbox exec <id> "ls /Users/"
# Expected: Error - path not accessible in sandbox
```

### Provider Errors
```bash
# Invalid API key (if testing)
# Temporarily corrupt config
mv ~/.config/ayo/ayo.json ~/.config/ayo/ayo.json.bak
echo '{"providers":{"anthropic":{"api_key":"invalid"}}}' > ~/.config/ayo/ayo.json

./ayo "test"
# Expected: Error - authentication failed

# Restore config
mv ~/.config/ayo/ayo.json.bak ~/.config/ayo/ayo.json
```

### Rate Limiting
```bash
# Send many requests quickly
for i in {1..10}; do ./ayo "count: $i" & done
wait

# Expected: All succeed or graceful rate limit handling
```

### Large File Handling
```bash
# Create large file
dd if=/dev/zero of=/tmp/large-file.bin bs=1M count=100

./ayo "Describe this file" -a /tmp/large-file.bin
# Expected: Graceful handling (error message or truncation warning)

rm /tmp/large-file.bin
```

### Concurrent Operations
```bash
# Start squad
./ayo squad start e2e-squad

# Try concurrent dispatches
./ayo "#e2e-squad" "task 1" &
./ayo "#e2e-squad" "task 2" &
./ayo "#e2e-squad" "task 3" &
wait

# Expected: All complete without race conditions
```

### Recovery from Corrupted State
```bash
# Corrupt a squad SQUAD.md
echo "invalid yaml ---" > ~/.local/share/ayo/sandboxes/squads/e2e-squad/SQUAD.md

./ayo squad show e2e-squad
# Expected: Error with helpful message

./ayo squad start e2e-squad
# Expected: Error - invalid SQUAD.md

# Recovery: restore valid SQUAD.md
# (use backup or recreate)
```

### Daemon Crash Recovery
```bash
# Find daemon PID
./ayo sandbox service status

# Force kill
kill -9 <daemon-pid>

# Try operations
./ayo sandbox list
# Expected: Error - cannot connect

# Cleanup socket
rm -f ~/.local/share/ayo/daemon.sock

# Restart daemon
./ayo sandbox service start
# Expected: Works
```

### Verification Criteria
- [ ] Invalid commands show helpful errors
- [ ] Missing daemon is clearly reported
- [ ] Unknown resources have clear errors
- [ ] Circular dependencies prevented
- [ ] Permission errors handled gracefully
- [ ] Provider errors reported clearly
- [ ] Large files handled safely
- [ ] Concurrent operations don't race
- [ ] Corrupted state is recoverable

## Acceptance Criteria

- [ ] Section written in guide
- [ ] All error scenarios documented
- [ ] Expected error messages included
- [ ] Recovery procedures for each error
- [ ] Edge cases tested and documented
