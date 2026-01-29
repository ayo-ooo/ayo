# `ayo` - Agents You Orchestrate

`ayo` is a command-line tool for running AI agents that can execute tasks, use tools, and chain together via Unix pipes.

## Quick Start

```bash
# Install
go install github.com/alexcabrera/ayo/cmd/ayo@latest

# Start chatting (built-ins install automatically)
ayo

# Single prompt
ayo "help me debug this test"

# With file attachment
ayo -a main.go "review this code"
```

## Features

- **Agents**: AI assistants with custom prompts and tool access
- **Skills**: Reusable instruction sets following the [agentskills spec](https://agentskills.org)
- **Tools**: Execute shell commands, delegate tasks, track todos
- **Memory**: Persistent facts and preferences across sessions
- **Sessions**: Resume previous conversations
- **Chaining**: Compose agents via Unix pipes
- **Plugins**: Extend with community packages

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         ayo CLI                             │
├─────────────────────────────────────────────────────────────┤
│  @ayo (default agent)                                       │
│  ├── bash tool      Execute shell commands                  │
│  ├── agent_call     Delegate to other agents                │
│  ├── todo           Track multi-step tasks                  │
│  └── skills         Instruction sets (ayo, debugging, etc.) │
├─────────────────────────────────────────────────────────────┤
│  Sessions           Persist conversation history            │
│  Memory             Store facts/preferences across sessions │
│  Plugins            Extend with community packages          │
├─────────────────────────────────────────────────────────────┤
│  Fantasy            Provider-agnostic LLM abstraction       │
│  (OpenAI, Anthropic, Google, OpenRouter)                    │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
~/.config/ayo/                    # User configuration
├── ayo.json                      # Main config file
├── agents/                       # User-defined agents
│   └── @myagent/
│       ├── config.json
│       ├── system.md
│       └── skills/
└── skills/                       # User-defined shared skills
    └── my-skill/
        └── SKILL.md

~/.local/share/ayo/               # Data and built-ins
├── ayo.db                        # Sessions and memories (SQLite)
├── agents/                       # Built-in agents (@ayo)
├── skills/                       # Built-in skills
└── plugins/                      # Installed plugins
```

## Built-in Agent

| Agent | Description |
|-------|-------------|
| `@ayo` | Default versatile assistant |

`@ayo` handles all tasks including agent/skill creation and management. Just ask:

```bash
ayo "help me create an agent for code review"
ayo "create a skill for debugging Go code"
```

## Documentation

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | Installation and first steps |
| [Agents](docs/agents.md) | Creating and managing agents |
| [Skills](docs/skills.md) | Extending agents with instructions |
| [Tools](docs/tools.md) | Tool system (bash, todo, etc.) |
| [Memory](docs/memory.md) | Persistent facts and preferences |
| [Sessions](docs/sessions.md) | Conversation persistence |
| [Chaining](docs/chaining.md) | Composing agents via pipes |
| [Delegation](docs/delegation.md) | Task routing to specialists |
| [Configuration](docs/configuration.md) | Config files and directories |
| [Plugins](docs/plugins.md) | Extending ayo |
| [CLI Reference](docs/cli-reference.md) | Complete command reference |

## Examples

### Create an Agent

Ask `@ayo` to help design your agent:

```bash
ayo "help me create an agent for code review"
```

Or use the CLI directly:

```bash
ayo agents create @reviewer \
  -m gpt-4.1 \
  -d "Reviews code for best practices" \
  -f ~/prompts/reviewer.md
```

### Install a Plugin

```bash
# Add research capabilities
ayo plugins install https://github.com/alexcabrera/ayo-plugins-research

# Use the new agent
ayo @research "latest developments in AI"
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

### Chain Agents

```bash
# Pipe output between agents with structured I/O
ayo @analyzer '{"code":"..."}' | ayo @reporter
```

## CLI Overview

```
ayo                              Start interactive chat
ayo "prompt"                     Run single prompt with @ayo
ayo @agent "prompt"              Run prompt with specific agent

ayo agents list                  List all agents
ayo agents create @name          Create new agent
ayo agents show @name            Show agent details

ayo skills list                  List all skills
ayo skills create name           Create new skill

ayo sessions list                List conversation sessions
ayo sessions continue            Resume a session

ayo memory list                  List memories
ayo memory search "query"        Search memories semantically

ayo plugins install <url>        Install plugin from git
ayo plugins list                 List installed plugins

ayo chain ls                     List chainable agents
ayo doctor                       Check system health
ayo setup                        Install/update built-ins
```

## Configuration

Config file: `~/.config/ayo/ayo.json`

```json
{
  "default_model": "gpt-4.1",
  "provider": { "name": "openai" }
}
```

Required: Set an API key environment variable:

```bash
export OPENAI_API_KEY="sk-..."
# or ANTHROPIC_API_KEY, OPENROUTER_API_KEY, GOOGLE_API_KEY
```

## System Health

```bash
ayo doctor      # Check dependencies and configuration
ayo doctor -v   # Verbose output with model list
```

## License

MIT
