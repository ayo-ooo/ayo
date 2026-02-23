# Ayo Documentation

Documentation for the ayo CLI framework.

## Getting Started

| Guide | Description |
|-------|-------------|
| **[Getting Started](getting-started.md)** | Install, configure, run your first agent |
| [Architecture](architecture.md) | System design and mental model |

## Core Guides

| Guide | Description |
|-------|-------------|
| [Agents](agents.md) | Creating and configuring agents |
| [Squads](squads.md) | Multi-agent team coordination |
| [Tickets](tickets.md) | Task coordination and dependencies |
| [Planners](planners.md) | Near-term todos and long-term planning |

## Features

| Guide | Description |
|-------|-------------|
| [Memory](memory.md) | Persistent knowledge across sessions |
| [Sessions](sessions.md) | Session management and continuity |
| [Skills](skills.md) | Reusable instruction modules |
| [Tools](tools.md) | Agent tool system |

## Reference

| Guide | Description |
|-------|-------------|
| [CLI Reference](cli-reference.md) | Complete command reference |
| [Configuration](configuration.md) | Config files and settings |

## Quick Reference

```bash
# Chat
ayo                         # Interactive with default agent
ayo "prompt"                # Single prompt
ayo @agent "prompt"         # Specific agent
ayo #squad "prompt"         # Dispatch to squad

# Agents
ayo agents list             # List all
ayo agents show @name       # Details
ayo agents create @name     # Create

# Squads
ayo squad create name       # Create with SQUAD.md
ayo squad list              # List all
ayo squad start name        # Start sandbox

# Tickets
ayo ticket create "title"   # Create ticket
ayo ticket list             # List tickets
ayo ticket ready            # Show ready work

# Sandbox
ayo sandbox list            # List sandboxes
ayo share ~/path            # Share directory
ayo sandbox shell @ayo      # Shell into sandbox

# System
ayo setup                   # Initial setup
ayo doctor                  # Health check
```
