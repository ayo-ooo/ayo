# Ayo Project Agent Memory

This file contains project-specific knowledge for AI agents working on the ayo codebase.

## Project Overview

Ayo is a CLI tool for running AI agents in sandboxed environments. Key components:

- **CLI** (`cmd/ayo/`): User-facing commands
- **Agents** (`internal/agent/`): Agent loading, configuration, identity
- **Sandbox** (`internal/sandbox/`): Docker container management for isolated execution
- **Daemon** (`internal/daemon/`): Background process for triggers, sessions, sandbox pool
- **Providers** (`internal/providers/`): LLM API integrations

## Architecture: Sandbox-First Execution

All agent commands execute inside an Alpine Linux container:

1. **Host process** handles LLM calls, memory, orchestration
2. **Container process** executes bash commands, file operations
3. **IRC server** (ngircd) enables inter-agent communication
4. **Mount system** provides controlled host filesystem access

## Debug Scripts

Located in `./debug/`, these scripts help diagnose issues:

### system-info.sh
Collects host system information.
```bash
./debug/system-info.sh           # Human-readable
./debug/system-info.sh --json    # Machine-readable
```
**Use when:** Starting any debugging session, reporting bugs.

### sandbox-status.sh
Reports sandbox container status and health.
```bash
./debug/sandbox-status.sh            # Quick check
./debug/sandbox-status.sh --verbose  # Include processes, IRC, users
```
**Use when:** Sandbox won't start, commands fail inside container.

### daemon-status.sh
Checks ayo daemon and background services.
```bash
./debug/daemon-status.sh         # Quick status
./debug/daemon-status.sh --logs  # Include recent logs
```
**Use when:** Triggers not firing, daemon won't connect.

### sandbox-exec.sh
Execute commands inside the sandbox for testing.
```bash
./debug/sandbox-exec.sh pwd                  # Run as root
./debug/sandbox-exec.sh --as ayo whoami      # Run as agent-ayo
./debug/sandbox-exec.sh --json cat /etc/os-release
```
**Use when:** Need to verify container state, test agent user permissions.

### irc-status.sh
Checks IRC server for inter-agent communication.
```bash
./debug/irc-status.sh                 # Quick check
./debug/irc-status.sh --messages 50   # More history
```
**Use when:** Agents can't communicate, IRC-related errors.

### mount-check.sh
Verifies mount permissions and file access.
```bash
./debug/mount-check.sh               # Check configured mounts
./debug/mount-check.sh --test-write  # Test write access
```
**Use when:** Agent can't access host files, permission errors.

### collect-all.sh
Comprehensive diagnostic report combining all scripts.
```bash
./debug/collect-all.sh                        # Print to stdout
./debug/collect-all.sh --output report.txt    # Save to file
./debug/collect-all.sh | pbcopy               # Copy for pasting
```
**Use when:** Creating bug reports, asking for help, need full context.

## Key Files

| Path | Purpose |
|------|---------|
| `cmd/ayo/root.go` | CLI entry point, global flags |
| `cmd/ayo/sandbox.go` | Sandbox management commands |
| `cmd/ayo/daemon.go` | Daemon control commands |
| `internal/sandbox/linux.go` | Docker container management |
| `internal/sandbox/apple.go` | macOS-specific sandbox (future) |
| `internal/daemon/server.go` | Daemon RPC server |
| `internal/daemon/trigger_engine.go` | Cron/watch trigger handling |
| `internal/providers/sandbox.go` | Sandbox provider interface |

## Common Debugging Workflows

### Agent command fails silently
1. Run `./debug/sandbox-status.sh --verbose`
2. Check container is running
3. Run `./debug/sandbox-exec.sh --as <agent> <failed-command>`
4. Look for permission or path errors

### Daemon won't start
1. Check existing process: `pgrep -f 'ayo daemon'`
2. Remove stale socket: `rm -f /tmp/ayo/daemon.sock`
3. Check logs: `./debug/daemon-status.sh --logs`
4. Restart: `ayo daemon start`

### Mount not working
1. Verify mount exists: `ayo mount list`
2. Check container mounts: `./debug/mount-check.sh`
3. Stop and restart sandbox
4. Re-add mount if needed

### Trigger not firing
1. Check daemon running: `ayo status`
2. List triggers: `ayo triggers list`
3. Test manually: `ayo triggers test <id>`
4. Check daemon logs: `./debug/daemon-status.sh --logs`

## Testing

Before submitting changes:
```bash
# Build
go build ./cmd/ayo/...

# Run tests
go test ./... -count=1

# Manual testing
# See MANUAL_TEST.md for comprehensive guide
```

## Code Conventions

- Use `globalOutput` from root.go for JSON/quiet flag support
- All CLI commands should support `--json` and `--quiet` (inherited from root)
- Sandbox operations go through `providers.SandboxProvider` interface
- Daemon communication uses JSON-RPC over Unix socket
- Background tasks use the trigger engine, not goroutines

## Memory File Locations

| File | Purpose |
|------|---------|
| `~/.config/ayo/ayo.json` | User configuration |
| `~/.local/share/ayo/ayo.db` | SQLite database (sessions, memory) |
| `~/.local/share/ayo/mounts.json` | Persistent mount permissions |
| `/tmp/ayo/daemon.sock` | Daemon Unix socket |
| `/tmp/ayo/daemon.pid` | Daemon PID file |
