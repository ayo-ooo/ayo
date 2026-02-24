# Implementation Analysis

Comprehensive implementation analysis of the ayo codebase based on actual code inspection and CLI behavior testing.

*Generated: 2026-02-24*

---

## CLI Commands

Based on `cmd/ayo/*.go` analysis and actual CLI testing.

### Core Commands

| Command | Description | Source File |
|---------|-------------|-------------|
| `ayo [prompt]` | Interactive mode or single prompt | `chat.go`, `root.go` |
| `ayo @agent [prompt]` | Chat with specific agent | `chat.go` |
| `ayo #squad [prompt]` | Send task to squad | `chat.go` |
| `ayo agent` | Agent management | `agents.go` |
| `ayo squad` | Squad management | `squad.go` |
| `ayo trigger` | Trigger management | `triggers.go` |
| `ayo memory` | Memory management | `memory.go` |
| `ayo plugin` | Plugin management | `plugins.go` |
| `ayo flow` | Flow execution | `flows.go` |
| `ayo service` | Service control (daemon) | `service.go` |
| `ayo ticket` | Ticket management | `tickets.go` |
| `ayo share` | Share management | `share.go` |
| `ayo backup` | Backup operations | `backup.go` |
| `ayo doctor` | Health check | `doctor.go` |
| `ayo setup` | Initial setup | `setup.go` |

### Global Flags

| Flag | Short | Description | Applies To |
|------|-------|-------------|------------|
| `--json` | | Output in JSON format | All commands |
| `--quiet` | `-q` | Minimal output | All commands |
| `--config PATH` | | Path to config file | All commands |
| `--no-jodas` | `-y` | Auto-approve file requests | All commands |
| `--help` | `-h` | Show help | All commands |
| `--debug` | | Show debug output | Root command |
| `--model` | `-m` | Override model | Root command |
| `--output` | `-o` | Target directory | Root command |

### ayo agent

```
ayo agent list [--json] [--quiet] [--trust LEVEL] [--type TYPE]
ayo agent create @name [--template TEMPLATE]
ayo agent show @name
ayo agent rm @name
ayo agent edit @name
ayo agent sessions
ayo agent wake @name
ayo agent sleep @name
```

**Flags:**
- `--trust`: Filter by trust level (`sandboxed`, `privileged`, `unrestricted`)
- `--type`: Filter by type (`builtin`, `user`)
- `--template`: Template for new agent (`default`, `reviewer`, `assistant`)

**Exit Codes:**
- `0`: Success
- `1`: Error (daemon not running, agent not found)

### ayo squad

```
ayo squad list [--json]
ayo squad create <name> [--from PATH] [--agents AGENTS]
ayo squad show <name>
ayo squad destroy <name> [--force]
ayo squad start <name>
ayo squad stop <name>
ayo squad add-agent <squad> <agent>
ayo squad remove-agent <squad> <agent>
ayo squad shell <squad> [agent]
ayo squad ticket <squad> <command>
ayo squad schema <command>
```

**Sub-commands for `squad ticket`:**
- `list`, `show`, `create`, `update`, `start`, `close`, `reopen`, `assign`

### ayo trigger

```
ayo trigger list [--json]
ayo trigger schedule <agent> <schedule> [--prompt PROMPT] [--name NAME]
ayo trigger watch <path> <agent> [patterns...] [--recursive] [--debounce DURATION]
ayo trigger create <type> [--config CONFIG]
ayo trigger show <id>
ayo trigger rm <id> [--force]
ayo trigger enable <id>
ayo trigger disable <id>
ayo trigger test <id>
ayo trigger history <id> [--limit N]
ayo trigger types
```

### ayo memory

```
ayo memory list [--json] [--category CAT] [--agent AGENT] [--topic TOPIC]
ayo memory store <content> [--category CAT] [--topic TOPIC] [--agent AGENT]
ayo memory search <query> [--limit N] [--threshold FLOAT]
ayo memory show <id>
ayo memory forget <id>
ayo memory link <id1> <id2>
ayo memory merge [--dry-run]
ayo memory export <file>
ayo memory import <file>
ayo memory reindex
ayo memory stats
ayo memory topics
ayo memory clear [--force]
ayo memory migrate
```

### ayo service

```
ayo service start [--foreground]
ayo service stop
ayo service status
```

**Note:** `ayo daemon` is a hidden alias for backwards compatibility.

---

## Configuration System

### File Locations

| Location | Purpose |
|----------|---------|
| `~/.config/ayo/ayo.json` | Global config |
| `~/.config/ayo/agents/@name/ayo.json` | Agent-specific config |
| `./.config/ayo/ayo.json` | Project-local config |
| `SQUAD.md` frontmatter | Squad config |

### Config Merge Order

1. Built-in defaults
2. Global config (`~/.config/ayo/ayo.json`)
3. Project config (`./.config/ayo/ayo.json`)
4. Agent-specific config
5. CLI flags (highest priority)

### Agent Configuration Schema

From `internal/agent/agent.go`:

```go
type Config struct {
    Model             string            `json:"model"`
    SystemFile        string            `json:"system_file"`
    Description       string            `json:"description,omitempty"`
    AllowedTools      []string          `json:"allowed_tools,omitempty"`
    DisabledTools     []string          `json:"disabled_tools,omitempty"`
    ModelConfig       *ModelConfig      `json:"model_config,omitempty"`
    TrustLevel        TrustLevel        `json:"trust,omitempty"`
    Guardrails        *bool             `json:"guardrails,omitempty"`
    Skills            []string          `json:"skills,omitempty"`
    ExcludeSkills     []string          `json:"exclude_skills,omitempty"`
    IgnoreBuiltinSkills bool            `json:"ignore_builtin_skills,omitempty"`
    IgnoreSharedSkills  bool            `json:"ignore_shared_skills,omitempty"`
    Memory            MemoryConfig      `json:"memory,omitempty"`
    Delegates         map[string]string `json:"delegates,omitempty"`
    Sandbox           SandboxConfig     `json:"sandbox,omitempty"`
    Permissions       *PermissionsConfig `json:"permissions,omitempty"`
    Triggers          []TriggerConfig   `json:"triggers,omitempty"`
}
```

### Default Values

| Field | Default |
|-------|---------|
| `model` | `claude-sonnet-4-20250514` |
| `trust` | `sandboxed` |
| `guardrails` | `true` |
| `sandbox.enabled` | `true` |
| `sandbox.persist_home` | `true` |
| `sandbox.network` | `true` |
| `memory.enabled` | `false` |
| `memory.scope` | `agent` |

### Trust Levels

| Level | Description |
|-------|-------------|
| `sandboxed` | All operations in container (default) |
| `privileged` | Can access host via file_request |
| `unrestricted` | Full host access (dangerous) |

### Model Configuration

```json
{
  "model_config": {
    "temperature": 0.7,
    "max_tokens": 4096,
    "top_p": 1.0
  }
}
```

### Sandbox Configuration

```json
{
  "sandbox": {
    "enabled": true,
    "user": "agent",
    "persist_home": true,
    "languages": ["go", "python", "node"],
    "image": "ubuntu:22.04",
    "network": true,
    "resources": {
      "memory": "2G",
      "cpu_shares": 1024
    },
    "mounts": [
      {"source": "/host/path", "target": "/container/path", "readonly": false}
    ]
  }
}
```

### Memory Configuration

```json
{
  "memory": {
    "enabled": true,
    "scope": "agent",
    "formation_triggers": {
      "preference": true,
      "correction": true,
      "fact": true,
      "pattern": false
    },
    "retrieval": {
      "max_results": 10,
      "threshold": 0.7
    }
  }
}
```

### Permissions Configuration

```json
{
  "permissions": {
    "file_request": true,
    "auto_approve": false,
    "allowed_paths": ["/home/user/projects"],
    "denied_paths": ["/home/user/.ssh"]
  }
}
```

---

## Sandbox Behavior

### Providers

| Provider | Platform | Source |
|----------|----------|--------|
| `applecontainer` | macOS 26+ | `internal/sandbox/apple/` |
| `nspawn` | Linux | `internal/sandbox/nspawn/` |
| `dummy` | Testing | `internal/sandbox/dummy/` |

### File System Layout (Host)

```
~/.local/share/ayo/
├── daemon.sock           # Unix socket
├── daemon.pid            # PID file
├── daemon.log            # Daemon logs
├── jobs.db               # Trigger job store
├── packages.json         # Plugin registry
├── memory/
│   ├── index.db          # SQLite index
│   └── zettelkasten/     # Markdown files
├── sandboxes/
│   ├── pool/             # Warm sandbox pool
│   └── squads/           # Squad sandboxes
│       └── {name}/
│           ├── SQUAD.md
│           ├── .tickets/
│           ├── workspace/
│           └── .ayo/
└── sessions/             # Session history
```

### File System Layout (Inside Sandbox)

```
/
├── run/
│   └── ayod.sock         # In-sandbox daemon socket
├── workspace/            # Main working directory
├── home/
│   └── {agent}/          # Per-agent home directories
└── shared/               # Host-mounted directories
```

### ayod Protocol

Socket: `/run/ayod.sock`

**RPC Methods:**

| Method | Request | Response |
|--------|---------|----------|
| `UserAdd` | `{username, shell, dotfiles}` | `{}` |
| `Exec` | `{user, command, env, cwd, timeout}` | `{exit_code, stdout, stderr}` |
| `ReadFile` | `{path}` | `{content}` |
| `WriteFile` | `{path, content, mode}` | `{}` |
| `Health` | `{}` | `{status}` |

### file_request Flow

1. Agent calls `file_request` tool with path and content
2. Tool sends request to host daemon via sandbox RPC
3. Daemon checks approval cache
4. If not cached: prompts user in terminal
5. User approves or rejects
6. If approved: file written to host path
7. Result cached for session
8. Audit log entry created

---

## Agent System

### Loading Order

1. Check user agents: `~/.config/ayo/agents/`
2. Check installed plugins: `~/.local/share/ayo/plugins/*/agents/`
3. Check built-in agents
4. First match wins (user overrides plugin, plugin overrides builtin)

### Agent Directory Structure

```
@agent-name/
├── ayo.json              # Configuration
├── agent.md              # System prompt (or system.md)
├── tools/                # Custom tools
│   └── my-tool/
│       ├── tool.json
│       └── run.sh
└── skills/               # Skills
    └── SKILL.md
```

### System Prompt Injection Order

1. Base system prompt (from `ayo.json` or `agent.md`)
2. Guardrails prefix (if enabled)
3. Squad constitution (if in squad context)
4. Skills content
5. Memory context (recent/relevant memories)
6. Guardrails suffix (if enabled)

### Tool Enablement Logic

```
enabled_tools = all_tools - disabled_tools
if allowed_tools specified:
    enabled_tools = allowed_tools ∩ enabled_tools
```

---

## Squad System

### SQUAD.md Frontmatter

```yaml
---
name: backend-team
lead: "@architect"
input_accepts: "@architect"
agents:
  - "@developer"
  - "@reviewer"
  - "@tester"
planners:
  near_term: "ayo-todos"
  long_term: "ayo-tickets"
---
```

### Frontmatter Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | (required) | Squad identifier |
| `lead` | string | `@ayo` | Lead agent for routing |
| `input_accepts` | string | Lead agent | Who receives input |
| `agents` | []string | [] | Member agents |
| `planners.near_term` | string | `ayo-todos` | Session task tracking |
| `planners.long_term` | string | `ayo-tickets` | Persistent coordination |

### Constitution Injection

The SQUAD.md content (after frontmatter) is injected into all member agents' system prompts wrapped in:

```xml
<squad_context>
{SQUAD.md content}
</squad_context>
```

### Ticket-Based Coordination

Squad uses `.tickets/` directory with markdown files:

```yaml
---
id: ayo-abc1
status: open
assignee: "@developer"
deps: []
tags: [feature]
---
# Ticket Title

Description...
```

### Agent User Creation

Each agent gets its own Linux user inside the sandbox:
- Username: sanitized agent handle (e.g., `@code` → `code`)
- Home: `/home/{username}/`
- Shell: `/bin/bash`
- Dotfiles: Copied from agent config if specified

---

## Trigger System

### Trigger Types

| Type | Description | Config |
|------|-------------|--------|
| `cron` | Cron expression schedule | `schedule: "0 9 * * *"` |
| `interval` | Fixed time interval | `every: "5m"` |
| `daily` | Daily at specific time | `times: ["09:00"]` |
| `weekly` | Weekly on specific days | `days: ["monday"], times: ["09:00"]` |
| `monthly` | Monthly on specific days | `days_of_month: [1], times: ["09:00"]` |
| `once` | One-time execution | `at: "2024-01-15T10:00:00Z"` |
| `watch` | File system changes | `path: "/src", patterns: ["*.go"]` |

### Trigger Configuration Schema

From `internal/daemon/trigger_engine.go`:

```go
type TriggerConfig struct {
    Schedule         string   `json:"schedule,omitempty"`
    Path             string   `json:"path,omitempty"`
    Paths            []string `json:"paths,omitempty"`
    Patterns         []string `json:"patterns,omitempty"`
    Exclude          []string `json:"exclude,omitempty"`
    Recursive        bool     `json:"recursive,omitempty"`
    Events           []string `json:"events,omitempty"`
    At               string   `json:"at,omitempty"`
    Every            string   `json:"every,omitempty"`
    StartImmediately bool     `json:"start_immediately,omitempty"`
    Times            []string `json:"times,omitempty"`
    Days             []string `json:"days,omitempty"`
    DaysOfMonth      []int    `json:"days_of_month,omitempty"`
    Timezone         string   `json:"timezone,omitempty"`
    Debounce         string   `json:"debounce,omitempty"`
    Singleton        bool     `json:"singleton,omitempty"`
}
```

### Watch Events

- `create` - File created
- `modify` - File modified
- `delete` - File deleted
- `rename` - File renamed

### Persistence

Triggers persist to SQLite via `internal/daemon/job_store.go`:

```sql
CREATE TABLE scheduled_jobs (
    id TEXT PRIMARY KEY,
    name TEXT,
    type TEXT NOT NULL,
    agent TEXT NOT NULL,
    config TEXT NOT NULL, -- JSON
    prompt TEXT,
    enabled INTEGER DEFAULT 1,
    created_at INTEGER,
    updated_at INTEGER
);

CREATE TABLE job_runs (
    id TEXT PRIMARY KEY,
    job_id TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at INTEGER,
    finished_at INTEGER,
    output TEXT,
    error TEXT
);
```

---

## Plugin System

### Manifest Schema

From `internal/plugins/manifest.go`:

```json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "description": "Plugin description",
  "author": "Author Name",
  "repository": "https://github.com/...",
  "license": "MIT",
  "ayo_version": ">=0.2.0",
  
  "agents": ["agent-handle"],
  "skills": ["skill-name"],
  "tools": ["tool-name"],
  
  "delegates": {
    "task-type": "@agent"
  },
  
  "default_tools": {
    "alias": "tool-name"
  },
  
  "dependencies": {
    "binaries": ["binary-name"],
    "plugins": ["other-plugin"]
  },
  
  "providers": [{
    "name": "provider-name",
    "type": "memory|sandbox|embedding|observer",
    "entry_point": "path/to/plugin.so",
    "config": {}
  }],
  
  "planners": [{
    "name": "planner-name",
    "type": "near|long",
    "entry_point": "path/to/plugin.so",
    "config": {}
  }],
  
  "triggers": [{
    "name": "trigger-name",
    "category": "poll|push|watch",
    "entry_point": "path/to/plugin.so",
    "config_schema": {}
  }],
  
  "squads": [{
    "name": "squad-name",
    "description": "...",
    "path": "squads/name",
    "agents": ["@agent1", "@agent2"]
  }],
  
  "post_install": "scripts/post-install.sh"
}
```

### Resolution Order

1. User components (`~/.config/ayo/agents/`, `~/.config/ayo/tools/`)
2. Installed plugins (by installation order)
3. Built-in components

### Plugin Registry

Stored in `~/.local/share/ayo/packages.json`:

```json
{
  "plugins": {
    "plugin-name": {
      "version": "1.0.0",
      "path": "/path/to/plugin",
      "installed_at": "2024-01-15T10:00:00Z"
    }
  }
}
```

---

## Memory System

### SQLite Schema

From `internal/memory/zettelkasten/index.go`:

```sql
CREATE TABLE memory_index (
    id TEXT PRIMARY KEY,
    category TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    agent_handle TEXT,
    path_scope TEXT,
    content TEXT NOT NULL,
    embedding BLOB,
    confidence REAL DEFAULT 1.0,
    topics TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    last_accessed_at INTEGER,
    access_count INTEGER DEFAULT 0,
    supersedes_id TEXT,
    superseded_by_id TEXT,
    unclear INTEGER DEFAULT 0
);

CREATE VIRTUAL TABLE memory_fts USING fts5(
    id UNINDEXED,
    content,
    topics,
    content='memory_index',
    content_rowid='rowid'
);

CREATE TABLE note_links (
    id TEXT PRIMARY KEY,
    from_note_id TEXT NOT NULL,
    to_note_id TEXT NOT NULL,
    relationship TEXT,
    created_at INTEGER NOT NULL,
    UNIQUE(from_note_id, to_note_id)
);
```

### Zettelkasten Format

Markdown files in `~/.local/share/ayo/memory/zettelkasten/`:

```markdown
---
id: 20240115-abc123
category: fact
status: active
agent: "@ayo"
scope: global
topics: ["go", "testing"]
created: 2024-01-15T10:00:00Z
updated: 2024-01-15T10:00:00Z
confidence: 1.0
---

# Memory Title

Memory content here...

## Links

- [[20240110-def456]] Related memory
```

### Memory Categories

| Category | Description |
|----------|-------------|
| `preference` | User preferences (tools, styles) |
| `fact` | Factual information about user/project |
| `correction` | User corrections to agent behavior |
| `pattern` | Observed behavioral patterns |

### Memory Scopes

| Scope | Description |
|-------|-------------|
| `global` | All agents, all paths |
| `agent` | Specific agent only |
| `path` | Specific directory only |
| `squad` | Squad members only |
| `hybrid` | Agent + path combination |

### Embedding Providers

| Provider | Model | Dimensions |
|----------|-------|------------|
| Anthropic | voyage-3 | 1024 |
| OpenAI | text-embedding-3-small | 1536 |
| Local | (configurable) | varies |

### Search Flow

1. Query text → embedding via EmbeddingProvider
2. Vector similarity search using cosine distance
3. Results filtered by scope and status
4. FTS fallback for keyword matching
5. Results ranked by similarity score
6. Access count incremented for retrieved memories

---

## Daemon Architecture

### Components

| Component | Purpose | Source |
|-----------|---------|--------|
| `Server` | RPC handler, socket listener | `server.go` |
| `TriggerEngine` | Scheduler, file watcher | `trigger_engine.go` |
| `JobStore` | Trigger persistence | `job_store.go` |
| `SandboxManager` | Sandbox lifecycle | `sandbox.go` |
| `AgentInvoker` | Agent execution | `invoker.go` |

### Socket Locations

| Platform | Path |
|----------|------|
| Unix/macOS | `~/.local/share/ayo/daemon.sock` |
| Windows | `\\.\pipe\ayo-daemon` |

### RPC Protocol

JSON-RPC 2.0 over Unix socket, newline-delimited.

See `internal/daemon/protocol.go` for full method list.

### Error Codes

| Code | Name |
|------|------|
| -32700 | Parse Error |
| -32600 | Invalid Request |
| -32601 | Method Not Found |
| -32602 | Invalid Params |
| -32603 | Internal Error |
| -1001 | Sandbox Not Found |
| -1002 | Sandbox Exhausted |
| -1003 | Sandbox Timeout |
| -1004 | Daemon Shutting Down |

---

## Testing Notes

### Build & Test

```bash
go build ./cmd/ayo/...
go test ./... -count=1
golangci-lint run
```

### Test Coverage Targets

| Package | Target | Notes |
|---------|--------|-------|
| `internal/sandbox` | 70% | Provider implementations |
| `internal/squads` | 70% | Squad coordination |
| `internal/daemon` | 70% | Background service |

### Debug Scripts

| Script | Purpose |
|--------|---------|
| `debug/system-info.sh` | Host system info |
| `debug/sandbox-status.sh` | Container status |
| `debug/daemon-status.sh` | Service status |

---

*This document is based on actual code inspection and CLI testing as of 2026-02-24.*
