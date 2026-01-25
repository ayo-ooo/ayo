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
- **Tools**: Execute shell commands, delegate tasks, manage plans
- **Memory**: Persistent facts and preferences across sessions
- **Sessions**: Resume previous conversations
- **Chaining**: Compose agents via Unix pipes
- **Plugins**: Extend with community packages

## Built-in Agents

| Agent | Description |
|-------|-------------|
| `@ayo` | Default versatile assistant |
| `@ayo.agents` | Agent management |
| `@ayo.skills` | Skill management |

```bash
ayo agents list
```

## Documentation

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | Installation and first steps |
| [Agents](docs/agents.md) | Creating and managing agents |
| [Skills](docs/skills.md) | Extending agents with instructions |
| [Tools](docs/tools.md) | Tool system (bash, plan, etc.) |
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
# Interactive wizard
ayo agents create @myagent

# Non-interactive
ayo agents create @helper -n \
  --model gpt-4.1 \
  --system "You are concise and helpful."
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
# Pipe output between agents
ayo @analyzer '{"code":"..."}' | ayo @reporter
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
ayo doctor
```

## License

MIT
