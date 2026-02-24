---
id: ayo-e2e3
status: open
deps: [ayo-e2e2]
links: []
created: 2026-02-24T14:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-e2e1
tags: [gtm, documentation, testing, e2e]
---
# Task: E2E Section 1 - Build, Installation & Setup

## Summary

Write Section 1 of the E2E Manual Testing Guide covering building from source, initial setup, and daemon management.

## Content Requirements

### Build from Source
```bash
cd /path/to/ayo
go build -o ayo ./cmd/ayo/...
./ayo version

# Verify build
./ayo --help
```

### Initial Setup
```bash
./ayo setup

# Expected outputs:
# - Created ~/.config/ayo/ayo.json
# - Created ~/.local/share/ayo/ directories
# - Installed default agents
# - Installed default prompts
```

### Configuration Verification
```bash
# Check config exists
cat ~/.config/ayo/ayo.json

# Check data directories
ls -la ~/.local/share/ayo/

# Check default agents installed
ls ~/.local/share/ayo/agents/
```

### Daemon Management
```bash
# Start daemon
./ayo sandbox service start
# Expected: "Daemon started"

# Check status
./ayo sandbox service status
# Expected: "Running" with PID

# Verify socket
ls -l ~/.local/share/ayo/daemon.sock

# Test daemon connection
./ayo sandbox list
# Expected: Empty list or existing sandboxes
```

### Doctor Check
```bash
./ayo doctor

# Expected checks:
# ✓ Configuration valid
# ✓ Daemon running
# ✓ Container runtime available
# ✓ Provider configured
# ✓ Default agents installed

./ayo doctor -v  # Verbose for detailed info
```

### Verification Criteria
- [ ] Build succeeds without errors
- [ ] `ayo version` shows version info
- [ ] `ayo setup` creates config files
- [ ] Daemon starts successfully
- [ ] `ayo doctor` shows all green

## Acceptance Criteria

- [ ] Section written in guide
- [ ] All commands tested and verified
- [ ] Expected outputs documented
- [ ] Error scenarios noted
