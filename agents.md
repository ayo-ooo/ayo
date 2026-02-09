# Ayo Project Agent Memory

This file contains project-specific knowledge for AI agents working on the ayo codebase.

## Project Overview

Ayo is a CLI tool for running AI agents in sandboxed environments. Key components:

- **CLI** (`cmd/ayo/`): User-facing commands
- **Agents** (`internal/agent/`): Agent loading, configuration, identity
- **Sandbox** (`internal/sandbox/`): Container management for isolated execution (Apple Container on macOS, systemd-nspawn on Linux)
- **Daemon** (`internal/daemon/`): Background process for triggers, sessions, sandbox pool
- **Providers** (`internal/providers/`): LLM API integrations

## Architecture: Sandbox-First Execution

All agent commands execute inside an Alpine Linux container:

1. **Host process** handles LLM calls, memory, orchestration
2. **Container process** executes bash commands, file operations
3. **IRC server** (ngircd) enables inter-agent communication
4. **Mount system** provides controlled host filesystem access

### Sandbox Providers

**Important:** This project does NOT use Docker. Container isolation is provided by:

| Provider | Platform | Implementation |
|----------|----------|----------------|
| `apple` | macOS 26+ (Apple Silicon) | Apple Container (`container` CLI) |
| `systemd-nspawn` | Linux with systemd | systemd-nspawn |
| `none` | All platforms | No-op fallback |

The service auto-selects the appropriate provider based on platform availability.

**Never use Docker commands** (`docker`, `docker-compose`, etc.) when working on this codebase. Use the `container` CLI on macOS or `systemd-nspawn` on Linux.

## Sandbox CLI Commands

### Lifecycle Management

```bash
# List active sandboxes
ayo sandbox list

# Show sandbox details
ayo sandbox show <id>

# Start/stop sandboxes
ayo sandbox start <id>
ayo sandbox stop <id>

# Remove stopped sandboxes (with optional home dir cleanup)
ayo sandbox prune
ayo sandbox prune --homes  # Also clean agent home directories
```

### Execution

```bash
# Run command in sandbox
ayo sandbox exec <id> <command> [args...]

# Open non-interactive shell (for agent use)
ayo sandbox shell <id>

# Open interactive login shell (human use only)
ayo sandbox login [id]
```

### File Transfer

The working copy model keeps host files safe while giving agents full write access:

```bash
# Push file to sandbox
ayo sandbox push <id> <local-path> <container-path>

# Pull file from sandbox
ayo sandbox pull <id> <container-path> <local-path>

# Sync working copy back to host
ayo sandbox sync <id> <host-path>

# Show differences before syncing
ayo sandbox diff <id> <host-path>
```

### Multi-Agent Collaboration

```bash
# Add agent to existing sandbox
ayo sandbox join <id> <agent>

# List agents in sandbox
ayo sandbox users <id>
```

### Resource Monitoring

```bash
# Show resource usage statistics
ayo sandbox stats <id>

# View sandbox logs
ayo sandbox logs <id>
```

## Mount Commands

Mounts persist filesystem access grants across sessions:

```bash
# Grant access to a path
ayo mount add <path>           # Read-write access
ayo mount add <path> --ro      # Read-only access

# List all grants
ayo mount list
ayo mount list --json          # JSON output

# Remove access
ayo mount rm <path>            # Remove specific grant
ayo mount rm --all             # Remove all grants
```

## Trigger Commands

```bash
# List triggers
ayo triggers list

# Show trigger details
ayo triggers show <id>

# Add trigger (cron or watch)
ayo triggers add --cron "0 * * * *" --agent myagent --prompt "Check status"
ayo triggers add --watch /path/to/dir --agent myagent --prompt "File changed"

# Remove trigger
ayo triggers rm <id>

# Test trigger manually
ayo triggers test <id>

# Enable/disable
ayo triggers enable <id>
ayo triggers disable <id>
```

## Agent Identity System

Each agent runs as a dedicated Linux user inside the sandbox:

| Component | Description |
|-----------|-------------|
| `agent-<handle>` | Linux username inside container |
| `/home/<handle>` | Agent home directory (persistent) |
| `/shared` | Shared workspace (all agents can access) |
| `/workspace/<session>` | Session-specific working copy |

### Working Copy Model

When an agent accesses host files:

1. Files are copied to `/workspace/<session>/` inside sandbox
2. Agent has full read/write access to the copy
3. Original host files remain untouched
4. User explicitly syncs changes back with `ayo sandbox sync`

This ensures:
- Agents can't accidentally corrupt host files
- User reviews all changes before applying
- Multiple agents can work on same files safely

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
Checks ayo service and background services.
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
| `cmd/ayo/mount.go` | Mount management commands |
| `cmd/ayo/triggers.go` | Trigger management commands |
| `cmd/ayo/service.go` | Service control commands |
| `internal/sandbox/apple.go` | Apple Container provider (macOS 26+) |
| `internal/sandbox/linux.go` | systemd-nspawn provider (Linux) |
| `internal/sandbox/none.go` | No-op fallback provider |
| `internal/sandbox/pool.go` | Warm sandbox pool management |
| `internal/sandbox/workingcopy/` | Working copy sync implementation |
| `internal/daemon/server.go` | Daemon RPC server |
| `internal/daemon/trigger_engine.go` | Cron/watch trigger handling |
| `internal/providers/providers.go` | Provider interfaces |
| `internal/tools/filerequest/` | File request tool for agents |
| `internal/tools/publish/` | Publish tool for agents |

## Common Debugging Workflows

### Agent command fails silently
1. Run `./debug/sandbox-status.sh --verbose`
2. Check container is running
3. Run `./debug/sandbox-exec.sh --as <agent> <failed-command>`
4. Look for permission or path errors

### Daemon won't start
1. Check existing process: `pgrep -f 'ayo service'`
2. Remove stale socket: `rm -f /tmp/ayo/daemon.sock`
3. Check logs: `./debug/daemon-status.sh --logs`
4. Restart: `ayo sandbox service start`

### Mount not working
1. Verify mount exists: `ayo mount list`
2. Check container mounts: `./debug/mount-check.sh`
3. Stop and restart sandbox
4. Re-add mount if needed

### Trigger not firing
1. Check daemon running: `ayo sandbox service status`
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
