# Ayo Documentation

> **ayo** - Agents You Orchestrate

Ayo is a command-line framework for creating, managing, and orchestrating AI agents that operate within isolated sandbox environments.

## Start Here

| Guide | Description |
|-------|-------------|
| **[Tutorial](TUTORIAL.md)** | **Comprehensive guide: philosophy, architecture, and practice** |
| [Getting Started](getting-started.md) | Quick installation and first steps |

## Core Guides

| Guide | Description |
|-------|-------------|
| [Agents](agents.md) | Creating and managing AI agents |
| [Skills](skills.md) | Extending agents with domain-specific instructions |
| [Tools](tools.md) | Tool system (bash, memory, delegation) |
| [Memory](memory.md) | Persistent facts and preferences |
| [Sessions](sessions.md) | Conversation persistence and resumption |

## Composition & Workflows

| Guide | Description |
|-------|-------------|
| [Flows](flows.md) | Composable agent pipelines with shell or YAML |
| [Chaining](chaining.md) | Composing agents via Unix pipes |
| [Delegation](delegation.md) | Task routing to specialized agents |

## System & Configuration

| Guide | Description |
|-------|-------------|
| [Configuration](configuration.md) | Config files, directories, environment |
| [Plugins](plugins.md) | Extending ayo with community packages |
| [CLI Reference](cli-reference.md) | Complete command reference |

## Technical References

| Guide | Description |
|-------|-------------|
| [Flows Specification](flows-spec.md) | YAML flow schema reference |
| [Offline Architecture](offline-architecture.md) | Browser-based offline mode |

## Key Concepts

### Agents

Agents are AI assistants defined as directories with configuration and system prompts:

```
@reviewer/
├── config.json    # Model, tools, settings
└── system.md      # Behavior instructions
```

```bash
ayo @reviewer "review main.go for security issues"
```

### Sandbox Isolation

Agent commands execute in containers, isolated from your host system:

```bash
ayo sandbox list          # List active sandboxes
ayo share ~/Code/project  # Share directory with sandbox
```

### Skills

Skills are reusable instruction modules following the [agentskills spec](https://agentskills.org):

```bash
ayo skills list           # List available skills
ayo skills show debugging # Show skill details
```

### Memory

Semantic memory persists facts and preferences across sessions:

```bash
ayo memory store "I prefer TypeScript"
ayo memory search "programming preferences"
```

### Flows

Flows orchestrate multi-step workflows:

```bash
ayo flows run daily-summary '{"project": "."}'
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│  HOST: CLI → LLM calls, memory, orchestration               │
├─────────────────────────────────────────────────────────────┤
│  DAEMON: Sandbox pool, triggers, Matrix communication       │
├─────────────────────────────────────────────────────────────┤
│  SANDBOX: Command execution, file operations (isolated)     │
└─────────────────────────────────────────────────────────────┘
```

## Getting Help

```bash
# General help
ayo --help

# Command-specific help
ayo agents --help
ayo flows --help

# Check system health
ayo doctor
ayo doctor -v  # Verbose with model list
```

## Quick Reference

```bash
# Chat
ayo                         # Interactive with default agent
ayo "prompt"                # Single prompt
ayo @agent "prompt"         # Specific agent
ayo -a file.txt "analyze"   # With attachment
ayo -c "follow up"          # Continue session

# Agents
ayo agents list             # List all
ayo agents show @name       # Details
ayo agents create @name     # Create

# Flows
ayo flows list              # List all
ayo flows run name [input]  # Execute
ayo flows new name          # Create

# Sandbox
ayo sandbox list            # List sandboxes
ayo share ~/path            # Share directory
ayo sandbox exec cmd        # Execute command

# System
ayo setup                   # Initial setup
ayo doctor                  # Health check
ayo sandbox service start   # Start daemon
```
