# Ayo Project Agent Memory

This file contains project-specific knowledge for AI agents working on the ayo codebase.

## Project Overview

Ayo is a CLI framework for creating, managing, and orchestrating AI agents that operate within isolated sandbox environments. Key components:

| Component | Location | Purpose |
|-----------|----------|---------|
| CLI | `cmd/ayo/` | User-facing commands |
| Agents | `internal/agent/` | Agent loading, configuration, identity |
| Sandbox | `internal/sandbox/` | Container management (Apple Container / systemd-nspawn) |
| Daemon | `internal/daemon/` | Background service for pool, triggers, communication |
| Providers | `internal/providers/` | LLM API integrations via Fantasy |
| Memory | `internal/memory/` | Persistent knowledge storage |
| Flows | `internal/flows/` | Workflow execution engine |
| Share | `internal/share/` | Host directory sharing |

## Core Philosophy

Ayo extends Unix philosophy to agent-based computing:

1. **Separation of concerns**: Host handles thinking (LLM), sandbox handles doing (commands)
2. **Isolation by default**: All agent commands execute in containers
3. **Trust is explicit**: Permissions granted, not assumed
4. **Files as interface**: Agents are directories, configuration is declarative
5. **Composability**: Agents chain via pipes with JSON schemas

## Architecture: Sandbox-First Execution

```
┌─────────────────────────────────────────────────────────────┐
│                          HOST                                │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  ayo CLI → LLM calls, memory, orchestration           │  │
│  └───────────────────────────────────────────────────────┘  │
│                              │                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Daemon → sandbox pool, triggers, Matrix              │  │
│  └───────────────────────────────────────────────────────┘  │
│                              │                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  SANDBOX → command execution, file operations         │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Sandbox Providers

**Important:** This project does NOT use Docker. Container isolation is provided by:

| Provider | Platform | Implementation |
|----------|----------|----------------|
| `apple` | macOS 26+ (Apple Silicon) | Apple Container (`container` CLI) |
| `systemd-nspawn` | Linux with systemd | systemd-nspawn |
| `none` | All platforms | No-op fallback |

The service auto-selects the appropriate provider based on platform availability.

**Never use Docker commands** when working on this codebase.

## CLI Command Reference

### Core Commands

```bash
ayo                              # Interactive chat
ayo "prompt"                     # Single prompt
ayo @agent "prompt"              # Specific agent
ayo -a file.txt "analyze"        # With attachment
ayo -c "follow up"               # Continue session
```

### Sandbox Commands

```bash
ayo sandbox list                 # List active sandboxes
ayo sandbox show <id>            # Show details
ayo sandbox exec <id> <cmd>      # Execute command
ayo sandbox login [id]           # Interactive shell
ayo sandbox push <id> <src> <dst># Copy to sandbox
ayo sandbox pull <id> <src> <dst># Copy from sandbox
ayo sandbox sync <id> <host>     # Sync back to host
ayo sandbox diff <id> <host>     # Show differences
ayo sandbox stop <id>            # Stop sandbox
ayo sandbox prune                # Remove stopped
```

### Share Commands

Shares provide instant access to host directories at `/workspace/{name}`:

```bash
ayo share ~/Code/project         # Share directory
ayo share . --as myproject       # With custom name
ayo share list                   # List shares
ayo share rm project             # Remove share
ayo share rm --all               # Remove all
```

### Service Commands

```bash
ayo sandbox service start        # Start daemon
ayo sandbox service start -f     # Foreground mode
ayo sandbox service stop         # Stop daemon
ayo sandbox service status       # Check status
```

### Trigger Commands

```bash
ayo triggers list                # List triggers
ayo triggers show <id>           # Show details
ayo triggers add --cron "..." --agent @a --prompt "..."
ayo triggers add --watch PATH --agent @a --prompt "..."
ayo triggers rm <id>             # Remove
ayo triggers test <id>           # Test manually
ayo triggers enable <id>         # Enable
ayo triggers disable <id>        # Disable
```

### Flow Commands

```bash
ayo flows list                   # List flows
ayo flows show <name>            # Show details
ayo flows run <name> [input]     # Execute
ayo flows new <name>             # Create shell flow
ayo flows new <name> --yaml      # Create YAML flow
ayo flows validate <file>        # Validate
ayo flows history                # Run history
ayo flows replay <run-id>        # Replay run
```

## Agent Identity System

Each agent runs as a dedicated Linux user inside the sandbox:

| Component | Description |
|-----------|-------------|
| `agent-<handle>` | Linux username inside container |
| `/home/<handle>` | Agent home directory (persistent) |
| `/shared` | Shared workspace (all agents can access) |
| `/workspace/{name}` | Host directories via shares |

## Key Files

| Path | Purpose |
|------|---------|
| `cmd/ayo/root.go` | CLI entry point, global flags |
| `cmd/ayo/sandbox.go` | Sandbox management commands |
| `cmd/ayo/share.go` | Share management commands |
| `cmd/ayo/triggers.go` | Trigger management commands |
| `cmd/ayo/flows.go` | Flow management commands |
| `internal/share/share.go` | Share service implementation |
| `internal/sandbox/apple.go` | Apple Container provider |
| `internal/sandbox/linux.go` | systemd-nspawn provider |
| `internal/sandbox/pool.go` | Warm sandbox pool |
| `internal/daemon/server.go` | Daemon RPC server |
| `internal/daemon/trigger_engine.go` | Trigger handling |
| `internal/flows/flow.go` | Flow execution |
| `internal/memory/service.go` | Memory storage |
| `internal/agent/agent.go` | Agent loading/config |

## Debug Scripts

Located in `./debug/`:

| Script | Purpose |
|--------|---------|
| `system-info.sh` | Host system information |
| `sandbox-status.sh` | Container status/health |
| `daemon-status.sh` | Service status |
| `sandbox-exec.sh` | Execute in sandbox |
| `irc-status.sh` | IRC server status |
| `mount-check.sh` | Mount permissions |
| `collect-all.sh` | Full diagnostic report |

## Common Debugging Workflows

### Agent command fails
1. `./debug/sandbox-status.sh --verbose`
2. Check container running
3. `./debug/sandbox-exec.sh --as <agent> <cmd>`

### Daemon won't start
1. `pgrep -f 'ayo service'`
2. `rm -f ~/.local/share/ayo/daemon.sock`
3. `./debug/daemon-status.sh --logs`
4. `ayo sandbox service start`

### Share not working
1. `ayo share list`
2. `ls -la ~/.local/share/ayo/sandbox/workspace/`
3. Verify `/workspace/{name}` inside sandbox

### Trigger not firing
1. `ayo sandbox service status`
2. `ayo triggers list`
3. `ayo triggers test <id>`
4. `./debug/daemon-status.sh --logs`

## Testing

```bash
# Build
go build ./cmd/ayo/...

# Run tests
go test ./... -count=1

# Manual testing
# See MANUAL_TESTING.md for guide
```

## Code Conventions

- Use `globalOutput` from root.go for JSON/quiet flag support
- All CLI commands support `--json` and `--quiet` (inherited)
- Sandbox operations go through `providers.SandboxProvider` interface
- Daemon communication uses JSON-RPC over Unix socket
- Background tasks use trigger engine, not goroutines

## Memory File Locations

| File | Purpose |
|------|---------|
| `~/.config/ayo/ayo.json` | User configuration |
| `~/.local/share/ayo/ayo.db` | SQLite database |
| `~/.local/share/ayo/shares.json` | Share configuration |
| `~/.local/share/ayo/daemon.sock` | Daemon socket |
| `~/.local/share/ayo/daemon.pid` | Daemon PID |

## Documentation

For comprehensive documentation, see:
- **[Tutorial](docs/TUTORIAL.md)**: Full system walkthrough
- **[README.md](README.md)**: Quick start and overview
- **[docs/](docs/)**: Individual topic guides
