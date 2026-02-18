# Ayo Documentation

This directory contains detailed documentation for the ayo CLI framework.

**For AI coding agents:** Start with AGENTS.md in the repo root for quick reference, then dive into these docs as needed.

**For humans:** Start with [TUTORIAL.md](TUTORIAL.md) for a comprehensive walkthrough.

---

> **ayo** - Agents You Orchestrate

Ayo is a command-line framework for creating, managing, and orchestrating AI agents that operate within isolated sandbox environments.

## Start Here

| Guide | Description |
|-------|-------------|
| **[Tutorial](TUTORIAL.md)** | **Comprehensive guide: philosophy, architecture, and practice** |
| [Architecture](architecture.md) | Unified mental model and decision tree |
| [Getting Started](getting-started.md) | Quick installation and first steps |

## Core Guides

| Guide | Description |
|-------|-------------|
| [Agents](agents.md) | Creating and managing AI agents |
| [Skills](skills.md) | Extending agents with domain-specific instructions |
| [Tools](tools.md) | Tool system (bash, memory, delegation) |
| [Memory](memory.md) | Persistent facts and preferences |
| [Sessions](sessions.md) | Conversation persistence and resumption |

## Multi-Agent Systems

| Guide | Description |
|-------|-------------|
| [**Squads**](squads.md) | **Team sandboxes with SQUAD.md constitutions** |
| [Tickets](tickets.md) | File-based coordination between agents |
| [Flows](flows.md) | Composable agent pipelines with shell or YAML |
| [I/O Schemas](io-schemas.md) | Structured I/O for agent pipelines |
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
| [CLI Reference](reference/README.md) | Complete command-line reference |

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

### Squads

Squads are isolated sandboxes where multiple agents collaborate. Each squad has a `SQUAD.md` constitution that defines the team's mission, roles, and coordination rules:

```
~/.local/share/ayo/sandboxes/squads/my-team/
├── SQUAD.md          # Team constitution (injected into all agents)
├── .tickets/         # Coordination tickets
├── workspace/        # Shared code workspace
└── agent-homes/      # Per-agent directories
```

```bash
ayo squad create my-team -a @backend,@frontend,@qa
$EDITOR ~/.local/share/ayo/sandboxes/squads/my-team/SQUAD.md
```

All agents in a squad receive the constitution in their system prompt, ensuring shared understanding of mission and coordination.

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

### Tickets

Tickets provide file-based coordination between agents:

```bash
ayo ticket create "Implement auth" -a @backend
ayo ticket create "Test auth" -a @qa --deps <auth-id>
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
│  DAEMON: Sandbox pool, triggers, squad management           │
├─────────────────────────────────────────────────────────────┤
│  SQUADS: Isolated team sandboxes with SQUAD.md constitution │
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
ayo squad --help
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

# Squads
ayo squad create name       # Create with SQUAD.md
ayo squad list              # List all
ayo squad start name        # Start sandbox
ayo squad stop name         # Stop sandbox

# Tickets
ayo ticket create "title"   # Create ticket
ayo ticket list             # List tickets
ayo ticket ready            # Show ready work

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
