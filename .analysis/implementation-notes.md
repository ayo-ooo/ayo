# Implementation Analysis

Analysis of ayo codebase for GTM documentation.

## CLI Commands

Based on cmd/ayo/*.go analysis:

### Core Commands
- `ayo` - Interactive mode or single prompt
- `ayo config` - Configuration management
- `ayo agent` - Agent management (list, show, create)
- `ayo squad` - Squad management (list, create, destroy, run)
- `ayo trigger` - Trigger management (list, add, remove, types)
- `ayo memory` - Memory management (list, add, search, export, import)
- `ayo plugin` - Plugin management (install, list, search)
- `ayo flow` - Flow execution
- `ayo daemon` - Daemon control (start, stop, status)
- `ayo share` - Share management
- `ayo backup` - Backup operations
- `ayo doctor` - Health check
- `ayo version` - Version info

### Global Flags
- `--json` - JSON output
- `--quiet` - Minimal output
- `--config` - Config file path
- `--no-jodas` - Auto-approve file requests

## Configuration

### ayo.json Schema (internal/agent/config.go)
- `provider` - LLM provider (anthropic, openai, vertex, ollama)
- `model` - Model name
- `tools` - Enabled tools
- `skills` - SKILL.md file paths
- `permissions` - File access settings
- `memory` - Memory configuration
- `planners` - Near-term/long-term planners

### Defaults
- Provider: anthropic
- Model: claude-sonnet-4-20250514
- Tools: All enabled by default

## Sandbox Behavior

### Providers (internal/sandbox/providers/)
- `applecontainer` - macOS 26+ Apple Container
- `nspawn` - Linux systemd-nspawn
- `dummy` - Testing stub

### File System Layout
```
~/.local/share/ayo/
├── sandboxes/
│   ├── ayo/           # Default sandbox
│   └── squads/        # Squad sandboxes
├── memory/            # Zettelkasten storage
├── plugins/           # Installed plugins
├── agents/            # User agents
├── triggers/          # Trigger configs
└── output/            # Agent output sync
```

### file_request Flow
1. Agent calls file_request tool
2. Host receives request via daemon RPC
3. User approves/rejects in terminal
4. On approve: file written to host path
5. Audit logged

## Agent System

### Loading (internal/agent/)
1. Check user agents dir
2. Check plugin agents
3. Check builtin agents
4. Merge configs (user > plugin > builtin)

### Directory Structure
```
@agent-name/
├── ayo.json          # Config
├── system.md         # System prompt
├── tools/            # Custom tools
└── skills/           # SKILL.md files
```

## Squad System

### SQUAD.md Processing (internal/squads/)
1. Parse YAML frontmatter
2. Extract agent roles
3. Build constitution
4. Inject into system prompts

### Coordination
- Ticket-based via .tickets/
- Dispatch routing by role
- Shared sandbox environment

## Trigger System

### Types (internal/daemon/triggers/)
- `cron` - Schedule-based (gocron v2)
- `interval` - Time interval
- `one_time` - Single execution
- `file_watch` - File system events

### Plugin Types (internal/triggers/)
- `poll` - Polling-based
- `push` - Event-driven
- `watch` - Continuous monitoring

## Plugin System

### Manifest Schema (internal/plugins/manifest.go)
```json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "components": {
    "agents": {},
    "tools": {},
    "skills": {},
    "squads": {},
    "triggers": {},
    "planners": {}
  }
}
```

### Resolution Order
1. User agents/tools
2. Installed plugins
3. Builtin components

## Memory System

### Storage (internal/memory/)
- SQLite index for semantic search
- Zettelkasten markdown files for human readability
- Embeddings via configured provider

### Scopes
- `global` - All agents, all paths
- `agent` - Specific agent only
- `path` - Specific directory only
- `squad` - Squad members only

### Categories
- `preference` - User preferences
- `fact` - Factual information
- `correction` - Behavior corrections
- `pattern` - Observed patterns

---

*Generated: 2026-02-23*
