# Ayo: Philosophy & Architecture

> **ayo** - Agents You Orchestrate

Ayo is a CLI tool for defining, managing, and running AI agents. It is the **execution engine** for agent workflows—not the orchestrator. Ayo is designed to be invoked by external systems (cron, Django background tasks, CI/CD, webhooks) that handle scheduling, triggers, and run management.

**The vision**: A terminal-native n8n in plaintext, agent-based rather than node-based. Ayo is the open-source instrument; a separate proprietary app (the "Bloomberg terminal for agent orchestration") will handle deployment, monitoring, and team management.

---

## Table of Contents

1. [Core Philosophy](#core-philosophy)
2. [Architecture Overview](#architecture-overview)
3. [Agent System](#agent-system)
4. [Skills System](#skills-system)
5. [Tool System](#tool-system)
6. [Session & Memory](#session--memory)
7. [Chaining & Structured I/O](#chaining--structured-io)
8. [Flows](#flows)
9. [Plugin System](#plugin-system)
10. [Configuration](#configuration)
11. [CLI Reference](#cli-reference)
12. [Design Principles](#design-principles)

---

## Core Philosophy

### What Ayo Is

- **A CLI tool** that runs AI agents from the terminal
- **An execution engine** that can be invoked by external orchestrators
- **A composition system** for piping agents together via Unix semantics
- **A runtime** with sessions, memory, and tool execution
- **Plaintext-first**: agents, skills, and flows are human-readable files

### What Ayo Is Not

- **Not a scheduler**: Use cron, systemd, or your orchestration layer
- **Not a webhook server**: Use your web framework or serverless functions
- **Not a GUI**: Terminal-native, scriptable, composable
- **Not a platform**: It's a tool that platforms can build on

### The Backronym

**A**gent **Y**ou **O**rchestrate

- **You** = the orchestration layer (Django, n8n, GitHub Actions, cron)
- **Orchestrate** = ayo is the thing being orchestrated
- **Agent** = the unit of work

Ayo is the instrument, not the conductor.

---

## Architecture Overview

### Directory Structure

```
/cmd/ayo/              # CLI entry points (Cobra commands)
/internal/             # Core packages (not importable externally)
  /agent/              # Agent loading, validation, schemas
  /run/                # Execution via Fantasy, tool orchestration
  /session/            # Session persistence
  /memory/             # Semantic memory with embeddings
  /skills/             # Skill discovery and parsing
  /plugins/            # Plugin installation and registry
  /config/             # Configuration loading
  /db/                 # SQLite database (sqlc generated)
  /paths/              # XDG-compliant path resolution
  /pipe/               # Stdin/stdout detection, chain context
  /ui/                 # Terminal rendering, spinners
  /builtin/            # Embedded agents, skills, prompts
/.read-only/           # Vendored reference implementations
```

### Key Packages

| Package | Responsibility |
|---------|----------------|
| `agent` | Agent definition, loading, validation, schema compatibility |
| `run` | Agent execution via Fantasy, tool orchestration, streaming |
| `session` | Session management and persistence |
| `memory` | Semantic memory storage, retrieval, embeddings |
| `skills` | Skill discovery, parsing, validation |
| `plugins` | Plugin installation, registry, manifest parsing |
| `config` | Configuration loading/saving |
| `db` | SQLite database access |
| `paths` | XDG-compliant path resolution |
| `pipe` | Stdin/stdout detection, chain context |
| `ui` | Terminal rendering, spinners, styled output |

### Data Flow

```
User Input
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  CLI (cmd/ayo)                                         │
│  - Parse args, load config                             │
│  - Resolve agent handle                                │
└─────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  Agent Loading (internal/agent)                        │
│  - Load config.json + system.md                        │
│  - Discover and compile skills                         │
│  - Build tool set                                      │
│  - Load input/output schemas                           │
└─────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  Runner (internal/run)                                 │
│  - Create Fantasy agent with tools                     │
│  - Stream response with callbacks                      │
│  - Handle tool execution                               │
│  - Persist messages to session                         │
└─────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  Fantasy (LLM abstraction)                             │
│  - Provider-agnostic API                               │
│  - Streaming, tool calling                             │
│  - Structured output                                   │
└─────────────────────────────────────────────────────────┘
    │
    ▼
Output (terminal or structured JSON)
```

---

## Agent System

### What Is an Agent?

An agent is a configured AI persona with:
- A **system prompt** defining its behavior
- A set of **tools** it can use
- A set of **skills** (domain knowledge)
- Optional **input/output schemas** for structured I/O
- Optional **memory** configuration
- Optional **delegates** for task handoff

### Agent File Structure

```
@agent-name/
├── config.json       # Agent configuration
├── system.md         # System prompt (Markdown)
├── input.jsonschema  # Optional: structured input validation
├── output.jsonschema # Optional: structured output schema
└── skills/           # Optional: agent-specific skills
    └── my-skill/
        └── SKILL.md
```

### config.json

```json
{
  "description": "A versatile command-line assistant",
  "model": "gpt-4.1",
  "allowed_tools": ["bash", "memory", "agent_call", "search"],
  "guardrails": true,
  "skills": ["coding", "debugging"],
  "exclude_skills": [],
  "ignore_builtin_skills": false,
  "ignore_shared_skills": false,
  "memory": {
    "enabled": true,
    "scope": "hybrid",
    "formation_triggers": {
      "on_correction": true,
      "on_preference": true,
      "on_project_fact": true,
      "explicit_only": false
    },
    "retrieval": {
      "auto_inject": true,
      "threshold": 0.3,
      "max_memories": 10
    }
  },
  "delegates": {
    "coding": "@crush",
    "research": "@ayo.research"
  }
}
```

### Agent Loading Priority

1. **Local project**: `./.config/ayo/agents/`
2. **User config**: `~/.config/ayo/agents/`
3. **User data**: `~/.local/share/ayo/agents/`
4. **Plugins**: `~/.local/share/ayo/plugins/*/agents/`

User agents shadow built-in agents with the same handle.

### System Prompt Assembly

The final system prompt is assembled from multiple sources:

1. **Prefix** (`~/.config/ayo/prompts/prefix.md`) - Global preamble
2. **Shared system** (`~/.config/ayo/prompts/system.md`) - Shared instructions
3. **Agent system** (`@agent/system.md`) - Agent-specific prompt
4. **Suffix** (`~/.config/ayo/prompts/suffix.md`) - Global postamble
5. **Tools prompt** - Dynamically generated tool documentation
6. **Skills prompt** - Compiled skill instructions in XML format
7. **Memory context** - Relevant memories injected at runtime
8. **Delegate context** - Available delegates and when to use them

### Built-in Agents

| Handle | Purpose |
|--------|---------|
| `@ayo` | Default agent, versatile CLI assistant |

The `@ayo` namespace is reserved—users cannot create `@ayo` or `@ayo.*` agents. Additional specialized agents can be added via plugins.

---

## Skills System

### What Is a Skill?

A skill is a package of domain-specific instructions that extends an agent's capabilities. Skills follow the [agentskills.org](https://agentskills.org) specification.

### Skill File Structure

```
skill-name/
├── SKILL.md          # Required: skill definition with YAML frontmatter
├── scripts/          # Optional: executable code
├── references/       # Optional: additional documentation
└── assets/           # Optional: templates, data files
```

### SKILL.md Format

```markdown
---
name: debugging
description: Systematic debugging techniques for identifying and fixing issues.
metadata:
  author: ayo
  version: "1.0"
---

# Debugging Skill

When debugging issues, follow these steps...
```

### Skill Discovery Priority

1. **Agent-specific**: `@agent/skills/skill-name/`
2. **User shared**: `~/.config/ayo/skills/skill-name/`
3. **Built-in**: `~/.local/share/ayo/skills/skill-name/`
4. **Plugin**: `~/.local/share/ayo/plugins/*/skills/skill-name/`

First match wins. Agent config can filter with `skills`, `exclude_skills`, `ignore_builtin_skills`, `ignore_shared_skills`.

### Built-in Skills

| Skill | Purpose |
|-------|---------|
| `ayo` | CLI usage documentation |
| `coding` | Code creation with delegate awareness |
| `debugging` | Systematic debugging techniques |
| `memory` | Memory tool usage guidelines |
| `flows` | Composable agent pipelines |
| `plugins` | Plugin management |
| `agent-discovery` | Finding appropriate agents |

---

## Tool System

### What Is a Tool?

A tool is a capability the agent can invoke during execution. Tools are defined in Go and registered with the Fantasy runtime.

### Built-in Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `bash` | Execute shell commands | `command`, `description`, `timeout_seconds`, `working_dir` |
| `memory` | Store/search/list/forget memories | `operation`, `content`, `query`, `id` |
| `todo` | Track multi-step task progress | `todos` (flat list with `content`, `active_form`, `status`) |
| `agent_call` | Delegate to sub-agents | `agent`, `prompt` |
| `search` | Web search (via plugin) | `query` |

### Tool Set Resolution

1. Agent's `allowed_tools` list specifies tool names or categories
2. `config.default_tools` maps categories to concrete tools
3. Plugin tools are discovered from installed plugins
4. Stateful tools (todo) have per-session SQLite storage

### Bash Tool

The bash tool is the primary execution mechanism. Agents use it for all system interaction unless a more specific tool exists.

```json
{
  "command": "git status",
  "description": "Checking repository status",
  "timeout_seconds": 30,
  "working_dir": "/path/to/project"
}
```

### Plan Tool

Enables agents to track complex, multi-step work with hierarchical structure:

```json
{
  "phases": [
    {
      "name": "Phase 1: Setup",
      "status": "completed",
      "tasks": [
        {
          "content": "Install dependencies",
          "active_form": "Installing dependencies",
          "status": "completed"
        }
      ]
    }
  ]
}
```

---

## Session & Memory

### Sessions

Sessions persist conversation history to SQLite, enabling:
- Resuming previous conversations
- Reviewing past interactions
- Chaining context across invocations

**Database location**: `~/.local/share/ayo/ayo.db`

**Schema**:
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    agent_handle TEXT NOT NULL,
    title TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    chain_context TEXT,      -- JSON: depth, source, source_description
    input_schema TEXT,       -- JSON Schema for input validation
    output_schema TEXT,      -- JSON Schema for output casting
    plan TEXT                -- JSON: current plan state
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,      -- user, assistant, system
    parts TEXT NOT NULL,     -- JSON array of message parts
    model TEXT,
    provider TEXT,
    created_at DATETIME
);

CREATE TABLE session_edges (
    parent_id TEXT NOT NULL,
    child_id TEXT NOT NULL,
    PRIMARY KEY (parent_id, child_id)
);
```

### Memory

Memories are persistent facts, preferences, and patterns that help agents provide contextual responses across sessions.

**How it works**:
1. Agents use a small local LLM (ministral-3:3b via Ollama) to detect memorable information
2. Memories are stored with vector embeddings (nomic-embed-text) for semantic search
3. Relevant memories are automatically retrieved and injected into system prompts
4. Agents can use the `memory` tool to search, store, list, or forget memories

**Memory Categories**:
| Category | Description |
|----------|-------------|
| `preference` | User preferences (tools, styles, communication) |
| `fact` | Facts about user or project |
| `correction` | User corrections to agent behavior |
| `pattern` | Observed behavioral patterns |

**Schema**:
```sql
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    category TEXT NOT NULL,
    embedding BLOB,           -- Vector embedding for semantic search
    agent_handle TEXT,        -- NULL for global memories
    path_prefix TEXT,         -- NULL for non-path-scoped memories
    supersedes TEXT,          -- ID of memory this replaces
    superseded_by TEXT,       -- ID of memory that replaced this
    forgotten_at DATETIME,    -- Soft delete timestamp
    created_at DATETIME,
    updated_at DATETIME
);
```

**Memory Scopes**:
- **Global**: Applies to all agents
- **Agent-scoped**: Applies only to specific agent
- **Path-scoped**: Applies to specific project/directory

---

## Chaining & Structured I/O

### Overview

Agents can be composed via Unix pipes when they have structured input/output schemas. The output of one agent becomes the input to the next.

```bash
ayo @code-reviewer '{"repo":".", "files":["main.go"]}' | ayo @issue-reporter
```

### Schema Files

- `input.jsonschema` - Validates input; agent only accepts JSON matching this schema
- `output.jsonschema` - Structures output; final response is cast to this format

### Pipeline Behavior

- **Stdin is piped** → Agent reads JSON from stdin
- **Stdout is piped** → UI goes to stderr, raw JSON goes to stdout
- Full UI (spinners, reasoning, tool calls) always visible on stderr

### Schema Compatibility

When piping agents:

| Tier | Description |
|------|-------------|
| `Exact` | Output schema identical to input schema |
| `Structural` | Output has all required fields of input (superset OK) |
| `Freeform` | Target agent has no input schema (accepts anything) |
| `None` | Incompatible schemas |

### Chain Context

When agents are chained, context is passed via environment variable:
- `AYO_CHAIN_CONTEXT` contains JSON with `depth`, `source`, and `source_description`
- Freeform agents receive a preamble describing the chain context

### Chain Commands

```bash
ayo chain ls                          # List chainable agents
ayo chain from @agent                 # Find downstream receivers
ayo chain to @agent                   # Find upstream sources
ayo chain validate @agent '<json>'    # Validate input against schema
ayo chain example @agent              # Generate example input
```

---

## Flows

### Overview

Flows are **shell scripts that compose agents into pipelines**. They are invoked by external orchestrators (Django, cron, GitHub Actions) and return structured output.

### What Ayo Provides

| Ayo Does | Orchestrator Does |
|----------|-------------------|
| Run agents | Decide *when* to run |
| Execute flows | Store run history |
| Validate I/O schemas | Route webhooks to flows |
| Stream output | Display dashboards |
| Return structured JSON | Manage schedules |
| Exit with status codes | Handle retries/alerts |

### Flow File Format

Single file, shell script with structured frontmatter:

```bash
#!/usr/bin/env bash
# ayo:flow
# name: code-review
# description: Review code and create GitHub issues
# input: input.jsonschema
# output: output.jsonschema

set -euo pipefail

INPUT="${1:-$(cat)}"

# Stage 1: Review code
REVIEW=$(echo "$INPUT" | ayo @code-reviewer)

# Stage 2: Create issues if findings exist
if echo "$REVIEW" | jq -e '.findings | length > 0' > /dev/null; then
  echo "$REVIEW" | ayo @issue-reporter
else
  echo '{"status": "clean", "findings": []}'
fi
```

### Directory Structure

```
~/.config/ayo/flows/
├── code-review.sh
├── daily-standup.sh
└── research-report.sh
```

Or with schemas:
```
~/.config/ayo/flows/
├── code-review/
│   ├── flow.sh
│   ├── input.jsonschema
│   └── output.jsonschema
```

### I/O Contract

For orchestrators to reliably invoke flows:

| Aspect | Contract |
|--------|----------|
| **Input** | JSON via argument or stdin |
| **Output** | JSON to stdout |
| **Logs** | Stderr for logs, spinners, UI |
| **Exit codes** | 0 = success, non-zero = failure |
| **Validation** | `--validate` flag for dry-run |
| **Timeout** | `--timeout <seconds>` |

### Planned CLI

```bash
ayo flows list                    # List available flows
ayo flows show <name>             # Show flow details + schemas
ayo flows run <name> [input]      # Run flow, JSON out
ayo flows run <name> --validate   # Dry-run, validate input only
ayo flows new <name>              # Scaffold new flow
ayo flows validate <path>         # Validate flow file
```

### No Triggers in Ayo

Triggers (schedules, webhooks, events) are **not** ayo's responsibility. The orchestration layer handles:
- When to invoke flows
- Webhook endpoints
- Cron schedules
- Event routing
- Retry logic
- Run history storage

Ayo is the instrument, not the conductor.

---

## Plugin System

### Overview

Plugins extend ayo with additional agents, skills, and tools. They are distributed via git repositories.

### Repository Naming

Plugin repositories must be named `ayo-plugins-<name>`:
- `ayo-plugins-crush` for the "crush" plugin
- `ayo-plugins-research` for a "research" plugin

### Plugin Structure

```
ayo-plugins-<name>/
├── manifest.json     # Required: plugin metadata
├── agents/           # Optional: agent definitions
│   └── @agent-name/
├── skills/           # Optional: shared skills
│   └── skill-name/
└── tools/            # Optional: external tools
    └── tool-name/
```

### manifest.json

```json
{
  "name": "crush",
  "version": "1.0.0",
  "description": "Crush coding agent for ayo",
  "author": "alexcabrera",
  "repository": "https://github.com/alexcabrera/ayo-plugins-crush",
  "agents": ["@crush"],
  "skills": ["crush-coding"],
  "tools": ["crush"],
  "delegates": {
    "coding": "@crush"
  },
  "dependencies": {
    "binaries": ["crush"]
  },
  "ayo_version": ">=0.2.0"
}
```

### External Tools

Plugins can provide external tools that wrap CLI commands:

```json
{
  "name": "my-tool",
  "description": "What this tool does",
  "command": "my-binary",
  "args": ["--flag"],
  "parameters": [
    {
      "name": "input",
      "description": "Input text",
      "type": "string",
      "required": true
    }
  ],
  "timeout": 60
}
```

### Plugin Commands

```bash
ayo plugins install <git-url>     # Install from git
ayo plugins install --local <path> # Install for development
ayo plugins list                   # List installed plugins
ayo plugins show <name>            # Show plugin details
ayo plugins update                 # Update all plugins
ayo plugins update <name>          # Update specific plugin
ayo plugins remove <name>          # Uninstall plugin
```

### Installation Locations

- Plugins: `~/.local/share/ayo/plugins/<name>/`
- Registry: `~/.local/share/ayo/packages.json`

---

## Configuration

### Global Config

**Location**: `~/.config/ayo/ayo.json`

```json
{
  "$schema": "./ayo-schema.json",
  "agents_dir": "~/.config/ayo/agents",
  "skills_dir": "~/.config/ayo/skills",
  "system_prefix": "~/.config/ayo/prompts/prefix.md",
  "system_suffix": "~/.config/ayo/prompts/suffix.md",
  "default_model": "gpt-4.1",
  "small_model": "ministral-3b",
  "ollama_host": "http://localhost:11434",
  "provider": {},
  "embedding": {},
  "delegates": {
    "coding": "@crush"
  },
  "default_tools": {
    "search": "tavily"
  }
}
```

### Directory Config

**Location**: `.ayo.json` in project root (or any parent directory)

```json
{
  "delegates": {
    "coding": "@crush",
    "research": "@ayo.research"
  },
  "model": "gpt-4.1",
  "agent": "@ayo"
}
```

### Path Resolution

Ayo follows XDG conventions with priority ordering:

| Priority | Config | Data |
|----------|--------|------|
| 1 | `./.config/ayo/` | `./.local/share/ayo/` |
| 2 | `~/.config/ayo/` | `~/.local/share/ayo/` |

**Dev mode**: When running from source checkout, built-in data uses `./.ayo/` instead of `~/.local/share/ayo/`.

### Key Paths

| Purpose | Path |
|---------|------|
| Config file | `~/.config/ayo/ayo.json` |
| Database | `~/.local/share/ayo/ayo.db` |
| User agents | `~/.config/ayo/agents/` |
| Built-in agents | `~/.local/share/ayo/agents/` |
| User skills | `~/.config/ayo/skills/` |
| Built-in skills | `~/.local/share/ayo/skills/` |
| Plugins | `~/.local/share/ayo/plugins/` |
| Prompts | `~/.config/ayo/prompts/` |

---

## CLI Reference

### Root Command

```bash
ayo [@agent] [prompt]           # Chat with agent
ayo [@agent] [prompt] -a file   # Include file attachment
ayo                             # Interactive chat with @ayo
```

### Agent Commands

```bash
ayo agents list                 # List available agents
ayo agents show @handle         # Show agent details
ayo agents create @handle       # Create new agent
ayo agents update               # Update built-in agents
```

### Skill Commands

```bash
ayo skills list                 # List available skills
ayo skills show <name>          # Show skill details
ayo skills create <name>        # Create new skill
ayo skills validate <path>      # Validate skill directory
```

### Session Commands

```bash
ayo sessions list               # List sessions
ayo sessions list -a @agent     # Filter by agent
ayo sessions show <id>          # Show session details
ayo sessions continue           # Resume session (picker)
ayo sessions continue <id>      # Resume specific session
ayo sessions delete <id>        # Delete session
```

### Memory Commands

```bash
ayo memory list                 # List all memories
ayo memory list -a @agent       # Filter by agent
ayo memory search <query>       # Semantic search
ayo memory show <id>            # Show memory details
ayo memory store <content>      # Store new memory
ayo memory forget <id>          # Soft delete memory
ayo memory stats                # Show statistics
ayo memory clear                # Clear all (with confirmation)
```

### Chain Commands

```bash
ayo chain ls                    # List chainable agents
ayo chain from @agent           # Find downstream receivers
ayo chain to @agent             # Find upstream sources
ayo chain validate @agent <json> # Validate input
ayo chain example @agent        # Generate example
```

### Plugin Commands

```bash
ayo plugins install <url>       # Install from git
ayo plugins list                # List installed
ayo plugins show <name>         # Show details
ayo plugins update              # Update all
ayo plugins remove <name>       # Uninstall
```

### System Commands

```bash
ayo setup                       # Install built-ins
ayo setup --force               # Overwrite modifications
ayo doctor                      # Check system health
ayo doctor -v                   # Verbose with model list
```

---

## Design Principles

### 1. Plaintext First

Everything is human-readable files:
- Agents are `config.json` + `system.md`
- Skills are `SKILL.md` with YAML frontmatter
- Flows are shell scripts with frontmatter
- No binary formats, no proprietary schemas

### 2. Unix Philosophy

- **Do one thing well**: Ayo runs agents, orchestrators schedule them
- **Composable**: Agents pipe together via stdin/stdout
- **Text streams**: JSON in, JSON out
- **Exit codes**: 0 = success, non-zero = failure

### 3. LLM-Friendly

- Shell scripts are universally understood by LLMs
- Markdown prompts are natural for LLMs to read and write
- JSON schemas are standard and well-supported
- Agents can create and modify other agents

### 4. Progressive Disclosure

- `ayo "hello"` just works (uses @ayo)
- `ayo @agent "prompt"` for specific agents
- `ayo @agent --model gpt-4 "prompt"` for overrides
- Full config for power users

### 5. Separation of Concerns

| Layer | Responsibility |
|-------|----------------|
| **Ayo** | Execution engine (run agents, manage sessions, tools) |
| **Orchestrator** | Scheduling, triggers, webhooks, run history |
| **UI** | Dashboards, team management, cost tracking |

### 6. Extensibility via Plugins

- Agents, skills, and tools can be distributed as plugins
- Git-based installation for version control
- Namespace isolation prevents conflicts
- Delegate mappings let plugins enhance core agents

### 7. Memory as a First-Class Citizen

- Agents learn and remember across sessions
- Semantic search for relevant context
- Scoped memories (global, agent, project)
- Explicit and implicit memory formation

### 8. Security by Default

- Guardrails enabled by default
- Tools require explicit allowlisting
- No network access without explicit tools
- Secrets via environment variables

---

## Future Directions

### Near-term

1. **Flows implementation** - Shell scripts with frontmatter, CLI commands
2. **Flow validation** - Schema compatibility checking
3. **Flow skill** - Let @ayo compose and run flows

### Medium-term

1. **MCP integration** - Model Context Protocol support
2. **Streaming improvements** - Better real-time output
3. **Cost tracking** - Token usage per session/agent

### Long-term

1. **Orchestration app** - The "Bloomberg terminal" for agent management
2. **Team features** - Shared agents, skills, memories
3. **Deployment** - Remote agent execution

---

## Appendix: LLM Integration

### Fantasy Library

Ayo uses the Fantasy library for provider-agnostic LLM integration:

```go
// Create provider and model
provider, _ := openrouter.New(openrouter.WithAPIKey(key))
model, _ := provider.LanguageModel(ctx, "anthropic/claude-3.5-sonnet")

// Create agent with tools
agent := fantasy.NewAgent(model,
    fantasy.WithSystemPrompt("You are helpful."),
    fantasy.WithTools(myTools...),
    fantasy.OnTextDelta(func(delta string) { fmt.Print(delta) }),
)

// Generate with streaming
result, _ := agent.Generate(ctx, fantasy.AgentCall{
    Prompt: "Hello",
    StopWhen: fantasy.FinishReasonIs(fantasy.FinishReasonEndTurn),
})
```

### Supported Providers

- OpenAI
- Anthropic
- Google (Gemini)
- OpenRouter
- Any OpenAI-compatible API

### Callbacks

Fantasy provides callbacks for real-time UI updates:
- `OnTextDelta` - Streaming text chunks
- `OnToolCall` - Tool invocation start
- `OnToolResult` - Tool completion
- `OnReasoning` - Reasoning/thinking content

---

*This document represents the current understanding of ayo as of January 2025. It should be updated as the project evolves.*
