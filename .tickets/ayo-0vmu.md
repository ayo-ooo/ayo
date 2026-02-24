---
id: ayo-0vmu
status: open
deps: []
links: []
created: 2026-02-23T22:16:25Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [cli, doctor]
---
# Improve ayo doctor command

Enhance the `ayo doctor` command to provide comprehensive system health checks with actionable fix suggestions.

## Context

`ayo doctor` is the first-line troubleshooting tool. This ticket makes it comprehensive and actionable.

## Current Checks

Basic checks that may exist:
- Daemon running
- Config directory exists

## New Checks to Add

### System Requirements

```
✓ macOS 26.0+ detected (Apple Container support)
✓ Go 1.22+ available
✓ git available
```

### Sandbox Provider

```
✓ Apple Container available
  Provider: apple-container
  Version: 1.0
```

### Daemon Status

```
✓ Daemon running (pid: 12345)
  Socket: ~/.local/share/ayo/daemon.sock
  Uptime: 2h 15m
  Active agents: 3
  Active triggers: 5
```

### Configuration

```
✓ Config directory exists
  Path: ~/.config/ayo/
  Agents: 5
  
⚠ No API key found for Anthropic
  Fix: Set ANTHROPIC_API_KEY environment variable
  Docs: https://docs.ayo.dev/setup#api-keys
```

### Trigger Engine

```
✓ Trigger engine running
  Scheduled jobs: 5
  Next trigger: health-check in 3m
```

### Sandboxes

```
✓ Sandboxes directory exists
  Path: ~/.local/share/ayo/sandboxes/
  @ayo: running
  #dev-team: running
  #research: stopped
```

### Common Issues

```
✓ No stale lock files
✓ Database not corrupted
✓ Socket permissions correct
```

## Output Format

```
ayo doctor

System Requirements
  ✓ macOS 26.0.1 (Apple Container support)
  ✓ Go 1.22.5
  ✓ git 2.43.0

Sandbox Provider
  ✓ apple-container 1.0

Daemon
  ✓ Running (pid 12345, uptime 2h 15m)

Configuration
  ✓ Config directory: ~/.config/ayo/
  ✓ 5 agents configured
  ⚠ ANTHROPIC_API_KEY not set
    Run: export ANTHROPIC_API_KEY=your-key

Triggers
  ✓ Engine running
  ✓ 5 scheduled jobs

Sandboxes
  ✓ 2 running, 1 stopped

─────────────────────────────
Summary: 10 passed, 1 warning, 0 errors

Run 'ayo doctor --fix' to attempt automatic fixes.
```

## --fix Flag

Attempt automatic fixes where safe:

```bash
ayo doctor --fix
# Attempting fixes...
# ✓ Cleaned up stale lock file
# ✓ Fixed socket permissions
# 
# Could not fix:
# ✗ ANTHROPIC_API_KEY not set (requires manual action)
```

## Implementation

```go
// cmd/ayo/doctor.go
type Check struct {
    Name     string
    Category string
    Run      func() CheckResult
    Fix      func() error  // nil if not auto-fixable
}

type CheckResult struct {
    Status  Status  // Pass, Warn, Fail
    Message string
    Fix     string  // Suggested fix
    DocsURL string  // Link to docs
}
```

## Files to Modify

1. **`cmd/ayo/doctor.go`** - Main command
2. **`internal/doctor/checks.go`** (new) - Check implementations
3. **`internal/doctor/fixes.go`** (new) - Auto-fix implementations

## Acceptance Criteria

- [ ] All check categories implemented
- [ ] Clear pass/warn/fail status
- [ ] Actionable fix suggestions
- [ ] Links to documentation
- [ ] --fix flag for auto-fixable issues
- [ ] JSON output with --json
- [ ] Summary at end

## Testing

- Test with healthy system (all pass)
- Test with missing API key (warning)
- Test with daemon not running (failure)
- Test --fix flag
- Test --json output
