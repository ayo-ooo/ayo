# Ayo Documentation

Welcome to the ayo documentation. 

> **ayo** - Agents You Orchestrate

Ayo is a command-line tool for running AI agents that can execute tasks, use tools, and chain together via Unix pipes. It is the **execution engine** for agent workflows—designed to be invoked by external orchestrators like Django, cron, GitHub Actions, or any system that can run shell commands.

**Start here**: [Philosophy & Architecture](philosophy.md) for the full vision.

## Quick Links

| Guide | Description |
|-------|-------------|
| [Philosophy & Architecture](philosophy.md) | Vision, design principles, and system overview |
| [Getting Started](getting-started.md) | Installation and first steps |
| [Agents](agents.md) | Creating and managing AI agents |
| [Skills](skills.md) | Extending agents with domain-specific instructions |
| [Tools](tools.md) | Tool system (bash, plan, external tools) |
| [Flows](flows.md) | Composable agent pipelines |
| [Memory](memory.md) | Persistent facts and preferences across sessions |
| [Sessions](sessions.md) | Conversation persistence and resumption |
| [Chaining](chaining.md) | Composing agents via Unix pipes |
| [Delegation](delegation.md) | Task routing to specialized agents |
| [Configuration](configuration.md) | Config files, directories, and environment |
| [Plugins](plugins.md) | Extending ayo with community packages |
| [CLI Reference](cli-reference.md) | Complete command reference |

## Design Documents

| Document | Description |
|----------|-------------|
| [Flows](design/flows.md) | Composable agent pipelines (specification) |
| [Flows Implementation](design/flows-implementation.md) | Phased implementation plan with milestones |
| [Flows Stories & Tasks](design/flows-stories.md) | Granular stories, tasks, and atomic work units |

## Concepts

### Agents

Agents are AI assistants with custom system prompts and tool access. Each agent is a directory containing configuration and instructions.

```bash
ayo @ayo "help me debug this test"
```

### Skills

Skills are reusable instruction sets that teach agents specialized tasks. They follow the [agentskills spec](https://agentskills.org).

```bash
ayo skills list
```

### Tools

Tools give agents the ability to take actions. The default tool is `bash` for executing shell commands.

```bash
# Agent uses bash tool to run commands
ayo "list all Go files in this directory"
```

### Sessions

Sessions persist conversation history, allowing you to continue previous conversations.

```bash
ayo sessions continue
```

### Memory

Memory stores facts and preferences about you that persist across sessions.

```bash
ayo memory store "I prefer TypeScript"
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                        User                             │
│                          │                              │
│                          ▼                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │                  ayo CLI                         │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────────────┐  │   │
│  │  │ Agents  │  │ Skills  │  │ Memory/Sessions │  │   │
│  │  └────┬────┘  └────┬────┘  └────────┬────────┘  │   │
│  │       │            │                │           │   │
│  │       ▼            ▼                ▼           │   │
│  │  ┌─────────────────────────────────────────┐   │   │
│  │  │           Fantasy (LLM Layer)           │   │   │
│  │  └─────────────────────────────────────────┘   │   │
│  │                      │                          │   │
│  │       ┌──────────────┼──────────────┐          │   │
│  │       ▼              ▼              ▼          │   │
│  │  ┌─────────┐   ┌──────────┐   ┌──────────┐    │   │
│  │  │ OpenAI  │   │ Anthropic│   │  Google  │    │   │
│  │  └─────────┘   └──────────┘   └──────────┘    │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

## Getting Help

```bash
# General help
ayo --help

# Command-specific help
ayo agents --help
ayo agents create --help

# Check system health
ayo doctor
```
