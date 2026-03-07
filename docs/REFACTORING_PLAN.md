# Ayo Build System Refactoring Plan

## Overview

This plan transforms Ayo from a CLI framework for managing agents into a pure build system for creating standalone executable agents and teams, removing all vestiges of the old framework architecture.

## Key Changes

### Command Renames
- `ayo init` → `ayo fresh` - Create new agent projects
- `ayo validate` → `ayo checkit` - Validate configuration

### New Commands
- `ayo add-agent` - Add additional agents to existing projects (single-agent or squad)

### Architecture Changes
- **REMOVE**: Sandbox containers, daemon service, centralized state management
- **KEEP**: Agent/squad execution logic, Fantasy framework integration, tool system
- **NEW**: Project-based configuration (config.toml for agents, team.toml for teams)

## Ticket Dependency Graph

```
[Priority 1 - Command Changes]
├── ayo-c30f: Rename init command to fresh
│   └── ayo-j5c2: Add agent command
│       └── ayo-5ob9: Update project structure conventions
└── ayo-ouv8: Rename validate command to checkit
    └── ayo-z9oo: Remove old framework CLI commands

[Priority 2 - Removals]
├── ayo-gn7f: Remove sandbox daemon infrastructure
│   └── ayo-jfqs: Refactor agent loading
│   ├── ayo-seqz: Refactor run package
│   └── ayo-9nqe: Keep and refactor doctor command
├── ayo-ujgk: Remove persistence layer
│   └── ayo-z9hi: Refactor squad system
│       └── ayo-gl1u: Update root.go
│           ├── ayo-3bw1: Keep and refactor chat command
│           └── ayo-62q7: Update documentation
└── ayo-z9oo: Remove old framework CLI commands
    └── ayo-jfqs: Refactor agent loading

[Priority 3 - Evaluation]
├── ayo-62q7: Update documentation
├── ayo-auds: Remove flows command if not project type
├── ayo-erxf: Remove capabilities and guardrails if unused
├── ayo-i48q: Keep audit command
├── ayo-6uvn: Remove builtin agents
├── ayo-7dik: Remove share and triggers packages
├── ayo-nrua: Remove pipe and integration packages
└── ayo-y5bf: Remove approval and delegates packages
```

## All Tickets (19 total)

### Priority 1 - Command Changes (2 tickets)
- **ayo-c30f**: Rename init command to fresh
- **ayo-ouv8**: Rename validate command to checkit

### Priority 2 - Core Removals & Refactoring (11 tickets)

**Removals:**
- **ayo-z9oo**: Remove old framework CLI commands (depends on: c30f, ouv8)
- **ayo-gn7f**: Remove sandbox daemon infrastructure
- **ayo-ujgk**: Remove persistence layer
- **ayo-7dik**: Remove share and triggers packages
- **ayo-y5bf**: Remove approval and delegates packages
- **ayo-nrua**: Remove pipe and integration packages

**Refactoring:**
- **ayo-jfqs**: Refactor agent loading for embedded config (depends on: z9oo, gn7f)
- **ayo-z9hi**: Refactor squad system for team projects (depends on: ujgk, z9oo)
- **ayo-seqz**: Refactor run package for local execution (depends on: gn7f, jfqs)
- **ayo-gl1u**: Update root.go for new execution model (depends on: jfqs, z9hi)
- **ayo-9nqe**: Keep and refactor doctor command (depends on: gn7f, ujgk)

### Priority 3 - New Features (1 ticket)
- **ayo-j5c2**: Add agent command (depends on: c30f, 5ob9)

### Priority 3 - Cleanup (5 tickets)
- **ayo-5ob9**: Update project structure conventions
- **ayo-62q7**: Update documentation for build system (depends on: z9oo, gn7f, gl1u)
- **ayo-auds**: Remove flows command if not project type
- **ayo-erxf**: Remove capabilities and guardrails if unused
- **ayo-6uvn**: Remove builtin agents
- **ayo-3bw1**: Keep and refactor chat command (depends on: jfqs, gl1u)
- **ayo-i48q**: Keep audit command

## Packages to REMOVE

### cmd/ayo/ commands (to be removed)
- `agents.go` - Agent CRUD management
- `agents_capabilities.go` - Agent capabilities
- `agents_lifecycle.go` - Agent lifecycle
- `sandbox.go` - Sandbox management
- `service.go` - Daemon service
- `triggers.go` - Trigger management
- `sessions.go` - Session management
- `setup.go` - Setup wizard
- `backup.go` - Backup of old framework data
- `sync.go` - Git sync of old framework
- `share.go` - Host directory sharing
- `skills.go` - Skills management
- `plugins.go` - Plugin management
- `memory.go` - Memory management
- `notifications.go` - Notification system
- `migrate.go` - Migration from old framework
- `index.go` - Index/capabilities
- `planner.go` - Planner management
- `squad.go` - Squad management
- `squad_shell.go` - Squad shell
- `tickets.go` - Ticket management

### internal/ packages (to be removed)
- `daemon/` - Background service with Unix socket RPC
- `sandbox/` - Container execution (Docker/Apple/Linux)
- `ayod/` - In-sandbox daemon
- `db/` - SQLite persistence for sessions, memories
- `session/` - Session persistence and JSONL format
- `share/` - Host directory sharing
- `triggers/` - Trigger interface and registry
- `approval/` - Approval caching
- `delegates/` - Delegation system
- `pipe/` - Unix pipe integration
- `integration/` - E2E tests for old framework
- `builtin/agents/` - Built-in agents (@ayo)
- `capabilities/` - Agent capabilities indexing
- `guardrails/` - (to be evaluated) Guardrails system

### cmd/ directory (to be removed)
- `cmd/ayod/` - In-sandbox daemon binary

## Packages to KEEP and REFACTOR

### cmd/ayo/ commands (keep and refactor)
- `root.go` - Main entry point, refactor to execute built agents
- `build.go` - Build standalone executables (already implemented)
- `init.go` → `fresh.go` - Initialize projects
- `validate.go` → `checkit.go` - Validate configurations
- `chat.go` - Interactive chat, refactor for built agents
- `doctor.go` - Health checks, refactor for build system
- `audit.go` - Audit logging
- `flows.go` - (to be evaluated) Keep as separate project type

### internal/ packages (keep and refactor)
- `build/` - Build system (new, keep)
- `build/types/` - Configuration types (new, keep)
- `run/` - Agent execution, refactor for local (no sandbox)
- `agent/` - Agent loading, refactor for embedded config
- `squads/` - Squad orchestration, refactor for team.toml
- `tools/` - Tool implementations (keep all)
- `providers/` - LLM providers (keep all)
- `ui/` - TUI components (keep all)
- `cli/` - Output handling (keep all)
- `config/` - Configuration (keep for build-time)
- `paths/` - Path utilities (keep all)
- `util/` - Utilities (keep all)
- `debug/` - Debugging (keep all)
- `version/` - Versioning (keep all)
- `testutil/` - Test utilities (keep all)
- `skills/` - Skill discovery (keep for project local)
- `prompts/` - Prompt loading (keep for project local)
- `planners/` - Planning strategies (keep as optional)
- `memory/` - Memory system (keep for agent-local)
- `smallmodel/` - Small model for memory (keep)
- `embedding/` - Embedding for search (keep)
- `ollama/` - Ollama integration (keep)
- `hitl/` - Human-in-the-loop (keep as optional)
- `flows/` - (to be evaluated) Flows system

## New Project Structure

### Single-Agent Project
```
myagent/
├── config.toml              # Agent configuration (TOML)
├── skills/                  # Agent-specific skills
│   └── custom/
│       └── SKILL.md
├── tools/                   # Custom Go tools
│   └── mytool.go
└── prompts/                 # Prompt templates
    └── system.md
```

### Multi-Agent Team Project
```
myteam/
├── team.toml                # Team configuration
├── agents/                  # Multiple agents in team
│   ├── agent1/
│   │   └── config.toml
│   └── agent2/
│       └── config.toml
└── workspace/               # Shared working directory
```

### Commands for Teams
- `ayo fresh myteam --template team` - Initialize team project
- `ayo add-agent ./myteam --name reviewer` - Add agent to team
- `ayo build myteam --team` - Build team executable
- `./myteam` - Run team

## Execution Model Changes

### Old Framework
```bash
# Install agents in framework
ayo agent create myreviewer

# Run through ayo
ayo @myreviewer "review this code"

# Sessions managed centrally
ayo sessions list
```

### New Build System
```bash
# Create as standalone project
ayo fresh myreviewer

# Build executable
cd myreviewer
ayo build .

# Run directly
./myreviewer "review this code"

# Data managed locally
./.myreviewer/data/
```

## Implementation Order

### Phase 1: Command Renames (Priority 1)
1. `tk start ayo-c30f` - Rename init to fresh
2. `tk start ayo-ouv8` - Rename validate to checkit
3. Update all references in code
4. `tk close ayo-c30f`
5. `tk close ayo-ouv8`

### Phase 2: Removals (Priority 2)
1. `tk start ayo-z9oo` - Remove old CLI commands
2. `tk start ayo-gn7f` - Remove sandbox infrastructure
3. `tk start ayo-ujgk` - Remove persistence layer
4. `tk start ayo-7dik` - Remove share/triggers
5. `tk start ayo-y5bf` - Remove approval/delegates
6. `tk start ayo-nrua` - Remove pipe/integration
7. Close all removal tickets

### Phase 3: Refactoring (Priority 2)
1. `tk start ayo-jfqs` - Refactor agent loading
2. `tk start ayo-z9hi` - Refactor squad system
3. `tk start ayo-seqz` - Refactor run package
4. `tk start ayo-9nqe` - Refactor doctor command
5. `tk start ayo-gl1u` - Update root.go
6. `tk start ayo-3bw1` - Refactor chat command
7. Close all refactoring tickets

### Phase 4: New Features (Priority 3)
1. `tk start ayo-5ob9` - Update project structure
2. `tk start ayo-j5c2` - Implement add-agent command
3. Close new feature tickets

### Phase 5: Cleanup (Priority 3)
1. `tk start ayo-auds` - Evaluate flows
2. `tk start ayo-erxf` - Evaluate capabilities
3. `tk start ayo-6uvn` - Remove builtin agents
4. `tk start ayo-62q7` - Update documentation
5. `tk start ayo-i48q` - Keep audit command
6. Close cleanup tickets

## Testing Strategy

After each phase:
1. `go build ./cmd/ayo` - Verify compilation
2. `go test ./...` - Run tests (update failing tests)
3. Manual testing of changed commands

Final verification:
1. Create test agent with `ayo fresh`
2. Build with `ayo build`
3. Run built binary
4. Test all CLI modes (structured, freeform, hybrid)
5. Verify documentation accuracy

## Dependencies Summary

| Ticket | Dependencies | Dependents |
|--------|--------------|-------------|
| ayo-c30f | None | z9oo, j5c2 |
| ayo-ouv8 | None | z9oo |
| ayo-z9oo | c30f, ouv8 | jfqs, z9hi, 62q7 |
| ayo-gn7f | None | jfqs, seqz, 9nqe, 62q7 |
| ayo-ujgk | None | z9hi, 9nqe |
| ayo-jfqs | z9oo, gn7f | gl1u, seqz, 3bw1 |
| ayo-z9hi | ujgk, z9oo | gl1u |
| ayo-seqz | gn7f, jfqs | - |
| ayo-gl1u | jfqs, z9hi | 3bw1, 62q7 |
| ayo-9nqe | gn7f, ujgk | - |
| ayo-3bw1 | jfqs, gl1u | - |
| ayo-j5c2 | c30f, 5ob9 | - |
| ayo-5ob9 | None | j5c2 |
| ayo-62q7 | z9oo, gn7f, gl1u | - |

## Risk Assessment

**High Risk:**
- Refactoring agent/squad loading - core execution path
- Removing persistence - affects multiple features
- Removing sandbox - requires new execution model

**Medium Risk:**
- Command renames - breaks user muscle memory
- Documentation updates - easy to miss references
- Removing old commands - may have hidden dependencies

**Low Risk:**
- Removing unused packages
- Updating help text
- Removing builtin agents

**Mitigation:**
- Thorough testing at each phase
- Keep old code as fallback during refactoring
- Update tests before removing code
- Comprehensive documentation updates
