# Architecture

Deep dive into ayo's system architecture for contributors and advanced users.

## Overview

Ayo follows a **host/sandbox architecture** where security-sensitive operations stay on the host while potentially dangerous code execution happens in isolated containers.

| Component | Location | Purpose |
|-----------|----------|---------|
| Host Process | Your machine | LLM calls, memory, orchestration, approvals |
| Sandbox Container | Isolated environment | Command execution, file operations |
| Daemon | Background service | Lifecycle management, triggers, coordination |

This separation ensures that AI agents can execute arbitrary code without compromising your system.

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│ Host Process                                                     │
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   CLI       │  │   Daemon    │  │   LLM Providers         │  │
│  │   (ayo)     │  │   (ayod)    │  │   (Anthropic, OpenAI)   │  │
│  └──────┬──────┘  └──────┬──────┘  └────────────┬────────────┘  │
│         │                │                      │               │
│         ▼                ▼                      ▼               │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                  Orchestration Layer                        │ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────────┐ │ │
│  │  │ Agent    │ │ Memory   │ │ Trigger  │ │ Squad          │ │ │
│  │  │ Loader   │ │ System   │ │ Engine   │ │ Dispatcher     │ │ │
│  │  └──────────┘ └──────────┘ └──────────┘ └────────────────┘ │ │
│  └──────────────────────────┬─────────────────────────────────┘ │
└─────────────────────────────┼───────────────────────────────────┘
                              │ JSON-RPC over Unix Socket
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ Sandbox Container                                                │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    ayod (in-sandbox daemon)                 │ │
│  │                                                             │ │
│  │  • Command execution (bash, tools)                          │ │
│  │  • File read/write operations                               │ │
│  │  • User management (per-agent users)                        │ │
│  │  • Environment setup                                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │ /workspace   │  │ /home/agent  │  │ Shared directories   │   │
│  │ (code)       │  │ (agent home) │  │ (mounted from host)  │   │
│  └──────────────┘  └──────────────┘  └──────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Core Subsystems

### CLI Layer

The `cmd/ayo/` directory contains all CLI commands:

| File | Commands |
|------|----------|
| `root.go` | Base command, global flags |
| `run.go` | `ayo`, `ayo run` - Agent execution |
| `squads.go` | `ayo squad *` - Squad management |
| `memory.go` | `ayo memory *` - Memory operations |
| `triggers.go` | `ayo trigger *` - Trigger management |
| `daemon.go` | `ayo service *` - Daemon control |
| `plugins.go` | `ayo plugin *` - Plugin management |

All commands inherit `--json` and `--quiet` flags via `globalOutput` for machine-readable output.

### Daemon Architecture

The daemon (`internal/daemon/`) is the central coordinator:

```go
// internal/daemon/server.go
type Server struct {
    listener     net.Listener
    triggerEng   *TriggerEngine
    jobStore     JobStore
    sandboxMgr   *SandboxManager
    invoker      AgentInvoker
}
```

**Responsibilities:**

| Component | Purpose |
|-----------|---------|
| `TriggerEngine` | Schedules and fires triggers |
| `JobStore` | Persists scheduled jobs to SQLite |
| `SandboxManager` | Manages sandbox lifecycle |
| `AgentInvoker` | Executes agents on demand |

**Event Loop:**

1. Accept RPC connection
2. Parse JSON-RPC request
3. Route to appropriate handler
4. Execute operation
5. Return JSON-RPC response

### RPC Protocol

Communication uses JSON-RPC 2.0 over Unix sockets:

**Socket locations:**
- Unix: `~/.local/share/ayo/daemon.sock`
- Windows: `\\.\pipe\ayo-daemon`

**Request format:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "agent.invoke",
  "params": {"agent": "code", "prompt": "Fix the bug"}
}
```

**Response format:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {"response": "...", "tokens_used": 1250}
}
```

See [RPC Reference](../reference/rpc.md) for complete method documentation.

### ayod Protocol

The in-sandbox daemon (`ayod`) runs as root inside containers:

**Socket:** `/run/ayod.sock` (inside sandbox)

**RPC Methods:**

| Method | Purpose |
|--------|---------|
| `UserAdd` | Create per-agent user accounts |
| `Exec` | Execute commands as specific user |
| `ReadFile` | Read file contents |
| `WriteFile` | Write file with permissions |
| `Health` | Health check |

**Request/Response types:**

```go
// internal/ayod/types.go
type ExecRequest struct {
    User    string            // Run as this user
    Command []string          // Command and arguments
    Env     map[string]string // Environment variables
    Cwd     string            // Working directory
    Timeout int               // Timeout in seconds
}

type ExecResponse struct {
    ExitCode int
    Stdout   string
    Stderr   string
}
```

### Sandbox Providers

Sandboxes are abstracted through the `SandboxProvider` interface:

```go
// internal/providers/providers.go
type SandboxProvider interface {
    Provider
    Create(ctx context.Context, opts SandboxCreateOptions) (Sandbox, error)
    Get(ctx context.Context, id string) (Sandbox, error)
    List(ctx context.Context) ([]Sandbox, error)
    Start(ctx context.Context, id string) error
    Stop(ctx context.Context, id string, opts SandboxStopOptions) error
    Delete(ctx context.Context, id string, force bool) error
    Exec(ctx context.Context, id string, opts ExecOptions) (ExecResult, error)
    Status(ctx context.Context, id string) (SandboxStatus, error)
    Stats(ctx context.Context, id string) (SandboxStats, error)
    EnsureAgentUser(ctx context.Context, id string, agentHandle string, dotfilesPath string) error
}
```

**Available providers:**

| Provider | Platform | Implementation |
|----------|----------|----------------|
| Apple Container | macOS 26+ | `internal/sandbox/apple/` |
| systemd-nspawn | Linux | `internal/sandbox/nspawn/` |

### LLM Integration

LLM providers implement the provider interface:

```go
// internal/providers/providers.go
type Provider interface {
    Name() string
    Type() ProviderType
    Init(ctx context.Context, config map[string]any) error
    Close() error
}
```

**Streaming:**

Responses stream token-by-token via channels:

```go
type StreamCallback func(delta StreamDelta)

type StreamDelta struct {
    Content   string
    ToolCalls []ToolCall
    Done      bool
}
```

**Supported providers:**

| Provider | Models |
|----------|--------|
| Anthropic | Claude 3.5, Claude 3 |
| OpenAI | GPT-4, GPT-3.5 |
| Google | Gemini |

### Memory Subsystem

Memory uses a hybrid storage approach:

```
~/.local/share/ayo/memory/
├── index.db          # SQLite: metadata, embeddings, search
└── zettelkasten/     # Markdown files: human-readable storage
    ├── 20240115-abc123.md
    └── ...
```

**MemoryProvider interface:**

```go
// internal/providers/providers.go
type MemoryProvider interface {
    Provider
    Create(ctx context.Context, m Memory) (Memory, error)
    Get(ctx context.Context, id string) (Memory, error)
    Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
    List(ctx context.Context, opts ListOptions) ([]Memory, error)
    Update(ctx context.Context, m Memory) error
    Forget(ctx context.Context, id string) error
    Supersede(ctx context.Context, oldID string, newMemory Memory, reason string) (Memory, error)
    Topics(ctx context.Context) ([]string, error)
    Link(ctx context.Context, id1, id2 string) error
    Unlink(ctx context.Context, id1, id2 string) error
    Reindex(ctx context.Context) error
}
```

**EmbeddingProvider interface:**

```go
type EmbeddingProvider interface {
    Provider
    Embed(ctx context.Context, text string) ([]float32, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
    Dimensions() int
    Model() string
}
```

**Search flow:**

1. Query text → embedding via EmbeddingProvider
2. Vector similarity search in SQLite
3. Results ranked by cosine similarity
4. Return matching Memory objects

### Trigger Engine

Triggers use gocron v2 for scheduling:

```go
// internal/daemon/trigger_engine.go
type TriggerEngine struct {
    scheduler gocron.Scheduler
    watcher   *fsnotify.Watcher
    jobStore  JobStore
    invoker   AgentInvoker
}
```

**Trigger types:**

| Type | Description | Config |
|------|-------------|--------|
| `cron` | Cron expression | `schedule: "0 9 * * *"` |
| `interval` | Fixed interval | `every: "5m"` |
| `daily` | Daily at time | `at: "09:00"` |
| `weekly` | Weekly | `days: ["Mon"], at: "09:00"` |
| `monthly` | Monthly | `days_of_month: [1], at: "09:00"` |
| `watch` | File changes | `path: "/src", patterns: ["*.go"]` |
| `once` | One-time | `at: "2024-01-15T10:00:00Z"` |

**JobStore interface:**

```go
type JobStore interface {
    Create(job *ScheduledJob) error
    Get(id string) (*ScheduledJob, error)
    List() ([]*ScheduledJob, error)
    Update(job *ScheduledJob) error
    Delete(id string) error
    RecordRun(run *JobRun) error
    UpdateRun(run *JobRun) error
    GetRecentRuns(jobID string, limit int) ([]*JobRun, error)
    LoadAllEnabled() ([]*ScheduledJob, error)
    Close() error
}
```

Jobs persist to SQLite, surviving daemon restarts.

### Planner System

Planners coordinate work across agents:

```go
// internal/planners/interface.go
type PlannerPlugin interface {
    Name() string
    Type() PlannerType  // NearTerm or LongTerm
    Init(ctx context.Context) error
    Close() error
    Tools() []fantasy.AgentTool
    Instructions() string
    StateDir() string
}
```

**Planner types:**

| Type | Scope | Example |
|------|-------|---------|
| `NearTerm` | Session | Todo lists, in-memory tasks |
| `LongTerm` | Persistent | Tickets, project management |

**Built-in planners:**

- `ayo-todos`: Session-scoped task tracking
- `ayo-tickets`: File-based ticket coordination

### Plugin System

Plugins are discovered and loaded at startup:

```
~/.config/ayo/plugins/
├── my-plugin/
│   ├── manifest.json
│   ├── agents/
│   ├── tools/
│   └── skills/
└── ...
```

**Resolution order:**

1. User components (`~/.config/ayo/`)
2. Installed plugins (`~/.local/share/ayo/plugins/`)
3. Built-in components

**Registration:**

```go
// internal/plugins/registry.go
type Registry struct {
    plugins map[string]*Plugin
    agents  map[string]*AgentRef
    tools   map[string]*ToolRef
    skills  map[string]*SkillRef
}
```

See [Plugin Reference](../reference/plugins.md) for manifest schema.

### Squad System

Squads are isolated team environments:

```
~/.local/share/ayo/sandboxes/squads/{name}/
├── SQUAD.md           # Team constitution
├── .tickets/          # Coordination tickets
├── workspace/         # Shared code
└── .ayo/              # Squad configuration
```

**Squad dispatch flow:**

1. Load SQUAD.md constitution
2. Identify ready tickets
3. Match tickets to capable agents
4. Wake agents with ticket context
5. Monitor completion
6. Handle handoffs between agents

**Constitution injection:**

```go
// internal/squads/constitution.go
func InjectConstitution(systemPrompt string, constitution *Constitution) string {
    // Inserts SQUAD.md content into agent's system prompt
}
```

## Data Flow

### Agent Invocation

```
User → CLI → Daemon → Agent Loader → LLM Provider
                 ↓
              Sandbox ← Tool Execution → ayod
                 ↓
              Response → Stream to User
```

### Tool Execution

```
LLM Response (tool call)
    ↓
Tool Router (internal/tools/)
    ↓
Sandbox Provider (if sandbox-bound)
    ↓
ayod.Exec() or ayod.ReadFile()/WriteFile()
    ↓
Result → Back to LLM
```

### file_request Flow

```
Agent requests file write
    ↓
file_request tool
    ↓
Check approval cache (internal/approval/)
    ↓
If not cached: Prompt user
    ↓
If approved: Write to host filesystem
    ↓
Log to audit trail
```

## Configuration Hierarchy

```
~/.config/ayo/
├── config.json         # Global configuration
├── agents/             # User-defined agents
├── skills/             # User-defined skills
└── plugins/            # Installed plugins

~/.local/share/ayo/
├── daemon.sock         # Daemon socket
├── daemon.pid          # Daemon PID
├── daemon.log          # Daemon logs
├── memory/             # Memory storage
│   ├── index.db
│   └── zettelkasten/
├── packages.json       # Plugin registry
├── jobs.db             # Scheduled jobs
└── sandboxes/          # Sandbox data
    ├── pool/           # Warm sandbox pool
    └── squads/         # Squad sandboxes
```

## See Also

- [Extending Ayo](extending.md) - Creating custom providers and plugins
- [Troubleshooting](troubleshooting.md) - Debugging common issues
- [RPC Reference](../reference/rpc.md) - Daemon API documentation
