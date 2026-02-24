# Core Concepts

This guide explains ayo's mental model. Understanding these concepts will help you use ayo effectively and build powerful agent workflows.

## Overview

Ayo is a CLI framework for running AI agents in isolated sandbox environments. The key principles are:

- **Isolation**: Agents execute commands inside containers, not on your host
- **Explicit Permission**: Host file modifications require your approval
- **Persistent Memory**: Agents remember context across sessions
- **Multi-Agent Coordination**: Squads enable team collaboration
- **Ambient Automation**: Triggers let agents act without prompting

## Agents

### What is an Agent?

An agent is an AI assistant configured for a specific task. Each agent has its own personality, capabilities, and access permissions defined in a directory structure.

### Agent Directory Structure

```
@agent-name/
├── config.json          # Model, tools, permissions
├── system.md            # Behavior instructions
├── input.jsonschema     # Input validation (optional)
├── output.jsonschema    # Output format (optional)
└── skills/              # Additional knowledge
    └── SKILL.md
```

### The Default @ayo Agent

Ayo ships with a built-in `@ayo` agent that serves as a general-purpose assistant. It runs in a dedicated sandbox and can:

- Execute shell commands
- Read and modify files (with approval)
- Store and retrieve memories
- Delegate to specialized agents

### Creating Custom Agents

```bash
ayo agents create @reviewer
```

This creates a directory at `~/.config/ayo/agents/@reviewer/`. Customize `system.md` to define the agent's behavior:

```markdown
You are a code review specialist. Focus on:
- Security vulnerabilities
- Performance issues  
- Code style consistency

Be direct and specific. Cite line numbers.
```

### Agent Loading Priority

When you invoke `@agent`, ayo searches in this order:

1. Local project: `./.config/ayo/agents/@agent/`
2. User agents: `~/.config/ayo/agents/@agent/`
3. System agents: `~/.local/share/ayo/agents/@agent/`
4. Installed plugins

The first match wins, allowing project-specific overrides.

### Trust Levels

Agents have three trust levels:

| Level | Sandbox | Host Access | Guardrails |
|-------|---------|-------------|------------|
| `sandboxed` (default) | Yes | Read-only mount | Yes |
| `privileged` | Yes | Read-only + file_request | Yes |
| `unrestricted` | No | Full host access | No |

Configure in `config.json`:

```json
{
  "trust_level": "sandboxed"
}
```

## Sandboxes

### What is a Sandbox?

A sandbox is an isolated container where agents execute commands. Sandboxes provide:

- **Security**: Agents cannot directly access your host filesystem
- **Consistency**: Clean environment for reproducible execution
- **Persistence**: Agent home directories survive restarts

### Sandbox Providers

Ayo supports two sandbox providers:

| Provider | Platform | Technology |
|----------|----------|------------|
| Apple Container | macOS 26+ | Native virtualization |
| systemd-nspawn | Linux | Namespace isolation |

### File System Layout

Inside a sandbox:

```
/
├── home/
│   └── {agent}/              # Agent's home directory
├── mnt/
│   └── {username}/           # Read-only host mount
├── output/                   # Safe write zone (syncs to host)
├── shared/                   # Shared between agents
└── workspaces/               # Project workspaces
```

On the host:

```
~/.local/share/ayo/
├── sandboxes/
│   ├── ayo/                  # @ayo's sandbox
│   └── squads/               # Squad sandboxes
│       └── {squad-name}/
├── output/                   # Synced from sandbox /output/
├── memory/                   # Zettelkasten storage
└── plugins/                  # Installed plugins
```

### The @ayo Sandbox

The default `@ayo` agent has a dedicated persistent sandbox. It's automatically created during setup and shared across sessions.

### Squad Sandboxes

Each squad gets its own isolated sandbox where all squad agents collaborate. Squad sandboxes include:

- Shared workspace at `/workspaces/{squad}/`
- Per-agent home directories
- Shared ticket system at `/.tickets/`

## Squads

### What is a Squad?

A squad is a team of agents working together in a shared sandbox. Squads enable complex workflows where multiple specialized agents collaborate on a task.

### Squad Directory Structure

```
~/.local/share/ayo/sandboxes/squads/{name}/
├── SQUAD.md                  # Team constitution
├── ayo.json                  # Squad configuration
├── workspace/                # Shared code workspace
├── .tickets/                 # Coordination tickets
└── .context/                 # Session persistence
    └── session.json
```

### SQUAD.md Constitution

The `SQUAD.md` file is the squad's constitution—it defines the mission, agent roles, and coordination rules. This document is injected into every agent's system prompt when they work in the squad.

```markdown
---
name: dev-team
planners:
  near_term: ayo-todos
  long_term: ayo-tickets
agents:
  - "@backend"
  - "@frontend"
---

# Squad: dev-team

## Mission
Build features with quality and speed.

## Agents

### @backend
- Implement API endpoints
- Write unit tests
- Handle database operations

### @frontend
- Build UI components
- Integrate with backend APIs
- Ensure responsive design

## Coordination
1. Break work into tickets
2. @backend implements API first
3. @frontend builds UI after API is ready
4. Both agents review each other's work
```

### Dispatch Routing

When you send a task to a squad, ayo routes it to the appropriate agent:

1. **Explicit target**: `ayo "#squad" @backend "implement endpoint"`
2. **Input matching**: Check agent `inputAccepts` patterns
3. **Squad lead**: Route to designated lead agent
4. **Default**: Route to `@ayo` for orchestration

### Coordination Through Tickets

Agents coordinate through markdown tickets in `/.tickets/`:

```markdown
---
id: feat-001
status: in_progress
assignee: "@backend"
deps: []
---
# Implement user authentication API

Create POST /api/auth/login endpoint...
```

When a ticket closes, dependent tickets automatically unblock.

## Memory

### What is Memory?

Memory is persistent knowledge that agents remember across sessions. Unlike conversation history (which is per-session), memories persist indefinitely and are retrieved based on relevance.

### Memory Categories

| Category | Description | Example |
|----------|-------------|---------|
| `preference` | User preferences | "I prefer TypeScript over JavaScript" |
| `fact` | Factual information | "The API endpoint is /api/v2/users" |
| `correction` | Behavior corrections | "Don't suggest semicolons in my Go code" |
| `pattern` | Observed patterns | "User usually asks about testing after implementation" |

### Memory Scopes

Memories can be scoped for different visibility:

| Scope | Visible To |
|-------|------------|
| `global` | All agents, everywhere |
| `agent` | Specific agent only |
| `path` | Agents working in a directory |
| `squad` | All agents in a squad |

### How Memory Works

1. **Storage**: Memories are embedded as vectors and stored in SQLite
2. **Retrieval**: When an agent starts, relevant memories are fetched via semantic search
3. **Context**: Retrieved memories are injected into the agent's context
4. **Formation**: Agents can store new memories via the `memory_store` tool

### Zettelkasten Files

Memories are also stored as human-readable markdown files:

```
~/.local/share/ayo/memory/
├── facts/
│   └── api-endpoint-v2.md
├── preferences/
│   └── typescript-preference.md
├── corrections/
│   └── no-semicolons-go.md
├── patterns/
│   └── testing-after-impl.md
└── .index.sqlite              # Search index
```

## Triggers

### What is a Trigger?

A trigger is an event that starts an agent automatically, enabling ambient automation without manual prompting.

### Trigger Types

| Type | Description | Example |
|------|-------------|---------|
| `cron` | Time-based schedule | "Every day at 9am" |
| `interval` | Recurring interval | "Every 30 minutes" |
| `once` | One-time execution | "Tomorrow at 2pm" |
| `watch` | File system changes | "When files in ./src change" |

### Cron Expressions

Ayo uses standard cron syntax with optional seconds:

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-6, Sun=0)
│ │ │ │ │
* * * * *
```

Aliases are also supported:

| Alias | Equivalent |
|-------|------------|
| `@hourly` | `0 * * * *` |
| `@daily` | `0 0 * * *` |
| `@weekly` | `0 0 * * 0` |
| `@monthly` | `0 0 1 * *` |

### File Watch Triggers

Watch triggers monitor directories for changes:

```bash
ayo triggers schedule code-reviewer \
  --watch ./src \
  --pattern "*.go" \
  --agent @reviewer \
  --prompt "Review changed Go files"
```

Features:
- **Glob patterns**: Filter which files trigger
- **Event types**: Create, modify, delete
- **Debouncing**: Batch rapid changes together
- **Singleton mode**: Prevent overlapping executions

## Tools

### What are Tools?

Tools are functions that agents can call to interact with the world. They're the bridge between AI reasoning and real actions.

### Built-in Tools

| Tool | Description |
|------|-------------|
| `bash` | Execute shell commands |
| `view` | Read file contents |
| `edit` | Modify files |
| `glob` | Find files by pattern |
| `grep` | Search file contents |
| `memory_store` | Save a memory |
| `memory_search` | Find relevant memories |
| `file_request` | Request host file access |
| `publish` | Write to /output (no approval needed) |
| `delegate` | Call another agent |
| `human_input` | Ask user for input |

### Tool Execution Context

Tools run in different contexts:

| Context | Tools | Description |
|---------|-------|-------------|
| `host` | memory, delegate | Run on your machine |
| `sandbox` | bash, view, edit | Run inside container |
| `bridge` | file_request, publish | Cross boundary |

### Enabling/Disabling Tools

Configure available tools in `config.json`:

```json
{
  "allowed_tools": ["bash", "view", "edit"],
  "disabled_tools": ["delegate"]
}
```

## Permissions

### The file_request Flow

When an agent needs to modify files on your host:

1. Agent calls `file_request` tool with path and content
2. Request appears in your terminal
3. You review and approve/deny
4. If approved, file is written to host

```
┌─────────────────────────────────────────────┐
│ @ayo wants to write:                        │
│   ~/Projects/app/main.go                    │
│                                             │
│ [Y]es  [N]o  [D]iff  [A]lways for session   │
└─────────────────────────────────────────────┘
```

### Approval Options

| Key | Action |
|-----|--------|
| `Y` | Approve this request |
| `N` | Deny and inform agent |
| `D` | View diff before deciding |
| `A` | Approve all similar requests this session |

### --no-jodas Mode

Skip approval prompts entirely:

```bash
ayo --no-jodas "refactor everything"
```

**Warning**: Use with caution. The agent can modify any file without asking.

### Permission Precedence

When determining if a request is auto-approved:

1. **Session cache**: Previous "Always" approval
2. **CLI flag**: `--no-jodas`
3. **Agent config**: `permissions.auto_approve`
4. **Global config**: Default approval settings

### Blocked Patterns

Some patterns are always blocked for security:

- `.git/*` - Git internals
- `.env*` - Environment files
- `**/secrets/*` - Secret directories
- `**/*.key`, `**/*.pem` - Private keys

## Plugins

### What is a Plugin?

A plugin is an extension that adds agents, tools, triggers, or other components. Plugins are distributed as directories or installable packages.

### Plugin Components

A plugin can provide:

| Component | Description |
|-----------|-------------|
| Agents | Pre-configured agents |
| Tools | External tools |
| Skills | Knowledge documents |
| Triggers | Event handlers |
| Planners | Task management |
| Squads | Pre-configured teams |

### Installing Plugins

```bash
ayo plugin install @acme/code-tools
ayo plugin install ./my-local-plugin
```

### Resolution Order

When loading components:

1. User-defined (`~/.config/ayo/`)
2. Installed plugins (`~/.local/share/ayo/plugins/`)
3. Built-in components

This allows you to override plugin behavior with custom configurations.

## Guardrails

### What are Guardrails?

Guardrails are safety prompts that constrain agent behavior. They're injected before and after the agent's system prompt.

### Prompt Injection Order

1. **System base prompt** - Core ayo behavior
2. **Guardrails prefix** - Safety constraints
3. **Agent system.md** - Agent personality
4. **Squad constitution** - Team rules (if in squad)
5. **Skills** - Additional knowledge
6. **Guardrails suffix** - Final constraints

### Customizing Guardrails

Override guardrails at `~/.config/ayo/prompts/guardrails/`:

```
prompts/
├── guardrails/
│   ├── default.md     # Default guardrails
│   └── @agent.md      # Agent-specific
└── sandwich/
    ├── prefix.md      # Before agent prompt
    └── suffix.md      # After agent prompt
```

---

*Ready to build? Continue to [Tutorials](tutorials/first-agent.md).*
