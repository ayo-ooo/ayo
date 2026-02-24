# Ayo Agent Memory

Quick reference for AI coding agents working on the ayo codebase.

---

## Ticket-Driven Development Workflow

**MANDATORY PROCESS** - Follow this exact workflow for all implementation work:

### Step 1: Select Next Ticket
1. Analyze ALL open tickets in `.tickets/`
2. Consider dependencies (tickets with satisfied deps first)
3. Consider priority (lower number = higher priority)
4. Select the single best ticket to work on next

### Step 2: Break Down into Atomic Tasks
1. Use the `todo` tool to load the ticket
2. Set first todo: "Analyze ticket and codebase"
3. Break the ticket into atomic work units (smallest possible tasks)
4. Each todo item = ONE discrete change (file, function, test)
5. Add final todo: "Run all tests and fix failures"

### Step 3: Implement
1. Work through todos one at a time
2. Mark each todo complete only when FULLY done
3. If tests fail, add new todos to analyze and fix
4. Continue until all todos complete and tests pass

### Step 4: Close Ticket
1. Change ticket status from `open` to `closed`
2. Verify all acceptance criteria met

### Step 5: Commit and Push
1. Stage all changes: `git add .`
2. Write extensive commit message explaining:
   - What was changed and why
   - Files modified
   - Any design decisions
   - Testing performed
3. Push to origin: `git push`
4. **Every ticket = separate commit** (no batching)

### Step 6: Compact and Continue
1. Use summarization tools to compact context window
2. Return to Step 1

### Rules
- **One ticket per commit** - Even rollup tickets, even interim fix tickets
- **No skipping steps** - Follow the workflow exactly
- **Atomic todos** - Break down until no further breakdown possible
- **Tests must pass** - Never close a ticket with failing tests
- **Push after every ticket** - Keep remote in sync

---

## Project Overview

Ayo is a CLI framework for creating, managing, and orchestrating AI agents that operate in isolated sandbox environments.

**Core architecture:**
- Host process: LLM calls, memory, orchestration
- Sandbox container: Command execution, file operations
- Squads: Isolated team sandboxes with SQUAD.md constitutions

**Sandbox providers:** Apple Container (macOS 26+), systemd-nspawn (Linux). NOT Docker.

For comprehensive documentation, see `PLAN.md` (docs/ will be created in Phase 9).

## Key References

| Topic | File | Purpose |
|-------|------|---------|
| GTM Plan | PLAN.md | Full implementation roadmap |
| Tickets | .tickets/*.md | All open/closed work items |
| This file | AGENTS.md | Agent memory and workflow |

## Key Directories

| Path | Purpose |
|------|---------|
| cmd/ayo/ | CLI entry points |
| cmd/ayod/ | In-sandbox daemon entry point |
| internal/agent/ | Agent loading, config, identity |
| internal/ayod/ | In-sandbox daemon (ayod) implementation |
| internal/sandbox/ | Container management, bootstrap |
| internal/squads/ | Squad management, SQUAD.md loading |
| internal/tickets/ | Ticket-based coordination |
| internal/planners/ | Planner plugin system (near/long-term) |
| internal/daemon/ | Background service |
| internal/providers/ | LLM API integrations |
| internal/memory/ | Persistent knowledge |
| internal/flows/ | Flow discovery and inspection |
| internal/share/ | Host directory sharing |
| internal/tools/ | Tool implementations |
| internal/ui/ | TUI components |

## Key Concepts

**ayod (In-Sandbox Daemon)**: Lightweight daemon running inside sandboxes:
- Runs as PID 1, manages sandbox lifecycle
- User management (creates agent users via `UserAdd`)
- Command execution as specific users (`Exec`)
- File read/write operations
- Communicates via JSON-RPC over `/run/ayod.sock`
- Binary: `cmd/ayod/`, implementation: `internal/ayod/`

**Squads**: Isolated sandboxes where multiple agents collaborate. Each squad has:
- `SQUAD.md` - Team constitution (mission, roles, coordination rules)
- `.tickets/` - File-based coordination
- `workspace/` - Shared code workspace
- Location: `~/.local/share/ayo/sandboxes/squads/{name}/`

**SQUAD.md**: Constitution file injected into all agents' system prompts. Defines mission, agent roles, coordination rules. See PLAN.md for details (docs/ created in Phase 9).

**Planners**: Pluggable modules for work coordination. Two types:
- Near-term (`ayo-todos`): Session-scoped task tracking
- Long-term (`ayo-tickets`): Persistent ticket-based coordination
- Interface: `internal/planners/interface.go` defines `PlannerPlugin`
- Each sandbox can have custom planners via config or SQUAD.md frontmatter

**Sandbox Bootstrap**: When a sandbox starts:
1. Copy ayod binary into container
2. Start ayod as PID 1
3. Create initial user (e.g., "ayo")
4. Set up standard directories (/workspace, /output, /mnt)

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
- **ayod not starting:** Check ayod binary exists with `make ayod`, verify container logs with `container logs <name>`
- **User creation fails:** Check ayod socket at `/run/ayod.sock` inside container
