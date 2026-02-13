# Ayo Agent Memory

Quick reference for AI coding agents working on the ayo codebase.

## Project Overview

Ayo is a CLI framework for creating, managing, and orchestrating AI agents that operate in isolated sandbox environments.

**Core architecture:**
- Host process: LLM calls, memory, orchestration
- Sandbox container: Command execution, file operations

**Sandbox providers:** Apple Container (macOS 26+), systemd-nspawn (Linux). NOT Docker.

For comprehensive documentation, see `docs/`.

## Documentation Map

| Topic | File | When to read |
|-------|------|--------------|
| Full system guide | docs/TUTORIAL.md | Deep understanding of architecture |
| Agent creation | docs/agents.md | Creating/modifying agents |
| Skills system | docs/skills.md | Adding knowledge to agents |
| Tools | docs/tools.md | Tool system, bash, memory |
| Flows | docs/flows.md | Multi-step workflows |
| CLI commands | docs/cli-reference.md | Command syntax |
| Configuration | docs/configuration.md | Config file locations |
| Plugins | docs/plugins.md | Extending ayo |

## Key Directories

| Path | Purpose |
|------|---------|
| cmd/ayo/ | CLI entry points |
| internal/agent/ | Agent loading, config, identity |
| internal/sandbox/ | Container management |
| internal/daemon/ | Background service |
| internal/providers/ | LLM API integrations |
| internal/memory/ | Persistent knowledge |
| internal/flows/ | Workflow execution |
| internal/share/ | Host directory sharing |
| internal/tools/ | Tool implementations |
| internal/ui/ | TUI components |

## Common Commands

**Build:** `go build ./cmd/ayo/...`

**Test:** `go test ./... -count=1`

**Run locally:** `go run ./cmd/ayo/... [args]`

**Lint:** `golangci-lint run`

## Code Conventions

- Use `globalOutput` from root.go for JSON/quiet flag support
- All CLI commands inherit `--json` and `--quiet` flags
- Sandbox operations use `providers.SandboxProvider` interface
- Daemon communication: JSON-RPC over Unix socket
- Background tasks use trigger engine, not goroutines
- This project does NOT use Docker

## Debugging

Scripts in `debug/`:
- `system-info.sh` - Host system information
- `sandbox-status.sh` - Container status
- `daemon-status.sh` - Service status

Common issues:
- **Daemon won't start:** `rm -f ~/.local/share/ayo/daemon.sock` then restart
- **Sandbox not working:** Check `ayo doctor` output
- **Share not mounting:** Verify path exists, check `ayo share list`
