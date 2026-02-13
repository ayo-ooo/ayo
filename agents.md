# Ayo Agent Memory

Quick reference for AI coding agents working on the ayo codebase.

## Project Overview

Ayo is a CLI framework for creating, managing, and orchestrating AI agents that operate in isolated sandbox environments.

**Core architecture:**
- Host process: LLM calls, memory, orchestration
- Sandbox container: Command execution, file operations
- Squads: Isolated team sandboxes with SQUAD.md constitutions

**Sandbox providers:** Apple Container (macOS 26+), systemd-nspawn (Linux). NOT Docker.

For comprehensive documentation, see `docs/`.

## Documentation Map

| Topic | File | When to read |
|-------|------|--------------|
| Full system guide | docs/TUTORIAL.md | Deep understanding of architecture |
| Agent creation | docs/agents.md | Creating/modifying agents |
| **Squads & SQUAD.md** | docs/squads.md | Multi-agent team coordination |
| Tickets | docs/tickets.md | File-based coordination |
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
| internal/squads/ | Squad management, SQUAD.md loading |
| internal/tickets/ | Ticket-based coordination |
| internal/daemon/ | Background service |
| internal/providers/ | LLM API integrations |
| internal/memory/ | Persistent knowledge |
| internal/flows/ | Workflow execution |
| internal/share/ | Host directory sharing |
| internal/tools/ | Tool implementations |
| internal/ui/ | TUI components |

## Key Concepts

**Squads**: Isolated sandboxes where multiple agents collaborate. Each squad has:
- `SQUAD.md` - Team constitution (mission, roles, coordination rules)
- `.tickets/` - File-based coordination
- `workspace/` - Shared code workspace
- Location: `~/.local/share/ayo/sandboxes/squads/{name}/`

**SQUAD.md**: Constitution file injected into all agents' system prompts. Defines mission, agent roles, coordination rules. See `docs/squads.md`.

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
- Squad constitutions loaded via `squads.LoadConstitution()` and injected via `squads.InjectConstitution()`

## Debugging

Scripts in `debug/`:
- `system-info.sh` - Host system information
- `sandbox-status.sh` - Container status
- `daemon-status.sh` - Service status

Common issues:
- **Daemon won't start:** `rm -f ~/.local/share/ayo/daemon.sock` then restart
- **Sandbox not working:** Check `ayo doctor` output
- **Share not mounting:** Verify path exists, check `ayo share list`
- **Squad RPC errors:** Restart daemon to pick up code changes
