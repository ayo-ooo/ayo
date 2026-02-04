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
- **Sandbox**: Isolated execution environments for secure command running
- **Daemon**: Background service for managing sandbox lifecycles

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
│  Sandbox            Isolated execution environments         │
│  Daemon             Background service for sandbox mgmt     │
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

Ask `@ayo` to help design your agent:

```bash
ayo "help me create an agent for code review"
```

Or use the CLI directly:

```bash
ayo agents create @reviewer \
  -m gpt-5.2 \
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

### Chat

```bash
ayo @ayo                         # Interactive chat with @ayo
ayo "prompt"                     # Single prompt with @ayo
ayo @agent "prompt"              # Single prompt with specific agent
ayo -a file.txt "analyze this"  # Attach file to prompt
```

### Agents

```bash
ayo agents list                  # List all agents
ayo agents show @name            # Show agent details
ayo agents create @name          # Create new agent
ayo agents update                # Update built-in agents
```

### Skills

```bash
ayo skills list                  # List all skills
ayo skills show <name>           # Show skill details
ayo skills create <name>         # Create new skill
ayo skills update                # Update built-in skills
```

### Sessions

```bash
ayo sessions list                # List conversation sessions
ayo sessions show <id>           # Show session details
ayo sessions continue            # Resume a session (interactive picker)
ayo sessions continue -l         # Resume most recent session
ayo sessions delete <id>         # Delete a session
```

### Memory

```bash
ayo memory list                  # List all memories
ayo memory search "query"        # Search memories semantically
ayo memory show <id>             # Show memory details
ayo memory store "content"       # Store a new memory
ayo memory forget <id>           # Forget a memory
ayo memory stats                 # Show memory statistics
```

### Flows

```bash
ayo flows list                   # List all flows
ayo flows show <name>            # Show flow details
ayo flows run <name> [input]     # Execute a flow
ayo flows new <name>             # Create new flow
ayo flows history                # Show run history
ayo flows replay <run-id>        # Replay a previous run
```

### Plugins

```bash
ayo plugins install <url>        # Install plugin from git
ayo plugins list                 # List installed plugins
ayo plugins show <name>          # Show plugin details
ayo plugins update               # Update all plugins
ayo plugins remove <name>        # Remove a plugin
```

### Chaining

```bash
ayo chain ls                     # List chainable agents
ayo chain inspect @agent         # Show agent schemas
ayo chain from @agent            # Find compatible downstream agents
ayo chain to @agent              # Find compatible upstream agents
ayo chain validate @agent <json> # Validate input against schema
ayo chain example @agent         # Generate example input
```

### Server & Web Client

```bash
ayo serve                        # Start HTTP API server
ayo serve --port 8080            # Start on specific port
ayo serve --host 0.0.0.0         # Allow external connections
```

Connect from the **[Web Client](https://alexcabrera.github.io/ayo-client-web/)** by scanning the QR code or entering the URL and token shown in the terminal.

### System

```bash
ayo setup                        # Install/update built-ins
ayo setup -f                     # Force reinstall
ayo doctor                       # Check system health
ayo doctor -v                    # Verbose with model list
ayo status                       # Show daemon and system status
```

### Daemon (Sandbox Management)

```bash
ayo daemon start                 # Start daemon in background
ayo daemon start --foreground    # Start in foreground (for debugging)
ayo daemon stop                  # Stop the daemon
```

The daemon manages sandbox lifecycles and provides IPC for sandbox operations.

## Configuration

Config file: `~/.config/ayo/ayo.json`

```json
{
  "default_model": "gpt-5.2",
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

## Sandbox Execution (Experimental)

Sandbox mode runs agent commands in isolated containers for security and reproducibility.

### Enabling Sandbox Mode

Configure an agent for sandbox execution in its `config.json`:

```json
{
  "sandbox": {
    "enabled": true,
    "provider": "docker",
    "image": "golang:1.22",
    "mounts": [
      {"host": ".", "container": "/workspace", "readonly": false}
    ],
    "network": true
  }
}
```

### Sandbox Providers

| Provider | Status | Requirements |
|----------|--------|--------------|
| Docker | Available | Docker installed |
| Lima | Available | Lima installed (macOS) |
| Apple Container | Planned | macOS 15+ |

### How It Works

1. Daemon starts (auto-started on first sandbox operation)
2. Agent requests sandbox via IPC
3. Container is created with configured mounts
4. Commands execute in the container
5. Sandbox is released when session ends

### Daemon Commands

```bash
ayo status                       # Check daemon status
ayo daemon start                 # Start daemon manually
ayo daemon stop                  # Stop daemon
```

## Offline Web Client (Experimental)

Try ayo directly in your browser with no installation required:

**[Launch ayo Offline Demo](https://alexcabrera.github.io/ayo/)**

The offline web client runs entirely in your browser using:
- **WebLLM** for local model inference (requires WebGPU)
- **TinyEMU** for Linux shell access via WebAssembly
- **IndexedDB** for persistent storage

### Features
- Chat with AI using local models or cloud API keys
- Full Linux terminal with BusyBox tools
- File editor for the browser-based filesystem
- Works offline after initial load

### Requirements
- Chrome 113+ or Edge 113+ (for WebGPU)
- 4GB+ available memory
- For local models: GPU with 2-8GB VRAM

See [Offline Mode User Guide](docs/user-guide.md) for details.
