# `ayo` - Agents You Orchestrate

`ayo` is a CLI framework for AI agents that actually do things. Create agents, teach them skills, let them execute commands in isolated sandboxes.

## Philosophy

Ayo extends the Unix philosophy to agent-based computing:

| Principle | Application |
|-----------|-------------|
| Do one thing well | Each agent has a focused purpose |
| Text streams as interface | JSON flows between agents via pipes |
| Small tools, composed | Simple agents combine into complex workflows |
| Files as universal abstraction | Agents are directories with configuration files |
| Isolation by default | Agents run in sandboxes, not on the host |
| Trust is explicit | Permissions are granted, not assumed |

**New to Ayo?** Start with the [Tutorial](docs/TUTORIAL.md) for a comprehensive introduction to the philosophy, architecture, and practice of agent-based computing.

## Quick Start

```bash
# Install
go install github.com/alexcabrera/ayo/cmd/ayo@latest

# Set API key
export ANTHROPIC_API_KEY="sk-..."

# Start chatting
ayo
```

## What Can You Do?

```bash
# Chat interactively
ayo

# Single task
ayo "help me debug this test"

# Review a file
ayo -a main.go "review this code"

# Continue a conversation
ayo -c "what about edge cases?"

# Use a specialized agent
ayo @reviewer "check for security issues"
```

## Features

| Feature | Description |
|---------|-------------|
| **Agents** | AI assistants with custom prompts and tool access |
| **Skills** | Reusable instruction sets following the [agentskills spec](https://agentskills.org) |
| **Tools** | Execute shell commands, delegate tasks, manage memory |
| **Memory** | Persistent facts and preferences across sessions |
| **Sessions** | Resume previous conversations |
| **Chaining** | Compose agents via Unix pipes with JSON schemas |
| **Delegation** | Route task types to specialist agents |
| **Sandbox** | Isolated execution environments (Apple Container / systemd-nspawn) |
| **Flows** | Multi-step workflows with shell scripts or YAML definitions |
| **Triggers** | Automated execution via cron, file watchers, webhooks |
| **Plugins** | Extend with community packages |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                          HOST                                │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  ayo CLI                                               │  │
│  │  • LLM API calls (Fantasy abstraction layer)          │  │
│  │  • Memory management (SQLite + embeddings)            │  │
│  │  • Session persistence                                 │  │
│  │  • Agent orchestration                                 │  │
│  └───────────────────────────────────────────────────────┘  │
│                              │                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Daemon (background service)                           │  │
│  │  • Sandbox pool management                             │  │
│  │  • Trigger engine (cron, watch, webhook)               │  │
│  │  • Inter-agent communication (Matrix/Conduit)          │  │
│  └───────────────────────────────────────────────────────┘  │
│                              │                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  SANDBOX CONTAINER                                     │  │
│  │  • Command execution (bash tool)                       │  │
│  │  • File operations                                     │  │
│  │  • Isolated from host filesystem                       │  │
│  │  • Per-agent home directories                          │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
~/.config/ayo/                    # Configuration
├── ayo.json                      # Main config
├── agents/                       # User-defined agents
│   └── @myagent/
│       ├── config.json
│       └── system.md
├── skills/                       # User-defined skills
├── flows/                        # User-defined flows
└── prompts/                      # prefix.md, suffix.md

~/.local/share/ayo/               # Data
├── ayo.db                        # SQLite (sessions, memory)
├── shares.json                   # Shared directories
├── agents/                       # Built-in agents
├── skills/                       # Built-in skills
├── plugins/                      # Installed plugins
└── sandbox/                      # Sandbox data
    ├── homes/                    # Agent home directories
    └── workspace/                # Host directory shares
```

## Documentation

| Guide | Description |
|-------|-------------|
| **[Tutorial](docs/TUTORIAL.md)** | **Comprehensive guide: philosophy, architecture, and practice** |
| [Getting Started](docs/getting-started.md) | Installation and first steps |
| [Agents](docs/agents.md) | Creating and managing agents |
| [Skills](docs/skills.md) | Extending agents with instructions |
| [Tools](docs/tools.md) | Tool system (bash, memory, delegation) |
| [Flows](docs/flows.md) | Composable agent pipelines |
| [Memory](docs/memory.md) | Persistent facts and preferences |
| [Sessions](docs/sessions.md) | Conversation persistence |
| [Chaining](docs/chaining.md) | Composing agents via pipes |
| [Delegation](docs/delegation.md) | Task routing to specialists |
| [Configuration](docs/configuration.md) | Config files and directories |
| [Plugins](docs/plugins.md) | Extending ayo |
| [CLI Reference](docs/cli-reference.md) | Complete command reference |

## Examples

### Create an Agent

```bash
# Interactive (recommended)
ayo "help me create an agent for code review"

# Direct CLI
ayo agents create @reviewer \
  -m claude-sonnet-4-20250514 \
  -d "Reviews code for best practices" \
  -f ~/prompts/reviewer.md
```

### Agent Chaining

```bash
# Type-safe pipeline with JSON schemas
ayo @analyzer '{"files":["main.go"]}' | ayo @reporter
```

### Configure Delegation

```bash
# Project-level config
cat > .ayo.json << 'EOF'
{
  "delegates": {
    "coding": "@crush"
  }
}
EOF

# Now @ayo delegates coding tasks automatically
ayo "refactor the auth module"
```

### Run a Flow

```bash
# Create and run a shell flow
ayo flows new daily-summary
ayo flows run daily-summary '{"project": "."}'
```

## CLI Overview

```bash
# Chat
ayo                              # Interactive with @ayo
ayo @agent "prompt"              # Single prompt
ayo -a file.txt "analyze"        # With attachment
ayo -c "follow up"               # Continue session

# Agents
ayo agents list                  # List agents
ayo agents show @name            # Show details
ayo agents create @name          # Create agent

# Flows
ayo flows list                   # List flows
ayo flows run name [input]       # Run flow
ayo flows new name               # Create flow

# Sandbox
ayo sandbox list                 # List sandboxes
ayo sandbox exec command         # Execute in sandbox
ayo share ~/Code/project         # Share directory

# Memory
ayo memory search "query"        # Semantic search
ayo memory store "fact"          # Store memory

# System
ayo setup                        # Initial setup
ayo doctor                       # Health check
ayo sandbox service start        # Start daemon
```

## Sandbox Providers

| Provider | Platform | Technology |
|----------|----------|------------|
| Apple Container | macOS 26+ (Apple Silicon) | Native containerization |
| systemd-nspawn | Linux with systemd | systemd containers |
| None | Fallback | Host execution (no isolation) |

The provider is auto-selected based on platform. Verify with `ayo doctor`.

## Configuration

**Main config**: `~/.config/ayo/ayo.json`

```json
{
  "default_model": "claude-sonnet-4-20250514",
  "provider": {"name": "anthropic"},
  "delegates": {
    "coding": "@crush"
  }
}
```

**Required**: Set an API key environment variable:

```bash
export ANTHROPIC_API_KEY="sk-..."
# or OPENAI_API_KEY, GOOGLE_API_KEY, OPENROUTER_API_KEY
```

## System Health

```bash
ayo doctor      # Check dependencies and configuration
ayo doctor -v   # Verbose output with model list
```

## License

MIT
