# ayo – Agents You Orchestrate

**ayo** is a CLI for managing AI agents that work in isolated sandbox environments. Create specialized agents, coordinate them through squads, and set up triggers for ambient automation.

```bash
# Just ask
ayo "help me debug this function"

# Use a specialized agent
ayo @reviewer "review this PR for security issues"

# Dispatch to a team
ayo "#dev-team" "build a user authentication system"
```

## Quick Start

```bash
# Install
go install github.com/alexcabrera/ayo/cmd/ayo@latest

# Configure
export ANTHROPIC_API_KEY="sk-ant-..."
ayo setup

# Start the daemon
ayo sandbox service start

# Start chatting
ayo
```

## Why ayo?

| Problem | ayo Solution |
|---------|--------------|
| Agents that just chat | Agents execute commands in isolated sandboxes |
| Lost context between sessions | Persistent memory and session resumption |
| Manual task management | Ticket-based coordination with dependencies |
| Security concerns | Sandboxed execution, explicit file access approval |
| Ad-hoc automation | Time and event-based triggers for ambient agents |

## Core Concepts

### Agents (`@name`)

Agents are AI assistants defined as directories:

```
@support/
├── config.json     # Model, tools, settings
└── system.md       # Behavior instructions
```

```bash
# Create an agent
ayo agents create @support --description "Customer support agent"

# Use it
ayo @support "help with this customer issue"
```

### Squads (`#team`)

Squads are isolated sandboxes where agents collaborate:

```bash
# Create a team
ayo squad create dev-team -a @frontend,@backend,@qa

# Start the squad
ayo squad start dev-team

# Dispatch work
ayo "#dev-team" "build the login feature"
```

Each squad has a `SQUAD.md` constitution defining roles and coordination:

```markdown
# Squad: dev-team

## Mission
Build features with quality and speed.

## Agents

### @backend
- Implement API endpoints
- Write tests

### @frontend
- Build UI components
- Integrate with backend

### @qa
- Review changes
- Write E2E tests

## Coordination
Use tickets to track work. @backend completes API first,
then @frontend implements UI, then @qa reviews.
```

### Tickets

Agents coordinate through markdown tickets:

```bash
# Create tickets
ayo ticket create "Implement login API" --assignee @backend
ayo ticket create "Build login form" --assignee @frontend --depends-on <api-ticket>

# See what's ready
ayo ticket ready

# Start and complete work
ayo ticket start <id>
ayo ticket close <id>
```

When a ticket closes, blocked dependents unblock automatically.

### Memory

Store persistent context that agents can access:

```bash
# Store facts
ayo memory store "This project uses PostgreSQL 15"
ayo memory store "Deploy to staging on Fridays"

# Search memories
ayo memory search "database"
```

Agents automatically retrieve relevant memories during conversations.

### Triggers

Set up ambient agents that act without prompting:

```bash
# Daily standup at 9am
ayo triggers schedule @standup "0 9 * * 1-5" \
  --prompt "Summarize yesterday's progress"

# Watch for file changes
ayo triggers watch ~/Projects @reviewer \
  --prompt "Review changed files" \
  --pattern "*.go"
```

See [Ambient Agent Patterns](docs/patterns/README.md) for common trigger setups.

### Sandbox Isolation

Agents run in containers with limited access:

- `/home/ayo/` – Agent's persistent home
- `/mnt/{username}/` – Read-only access to your files  
- `/output/` – Write freely (synced to host)

To write to your files, agents must request permission:

```
┌─────────────────────────────────────────────┐
│ @ayo wants to update:                       │
│   ~/Projects/app/main.go                    │
│ Reason: Fixed authentication bug            │
│                                             │
│ [Y]es  [N]o  [D]iff  [A]lways for session   │
└─────────────────────────────────────────────┘
```

## Architecture

```
┌────────────────────────────────────────────────────┐
│                    AYO CLI                          │
│  Direct invocation, squad dispatch, triggers       │
└────────────────────────────────────────────────────┘
                         │
┌────────────────────────────────────────────────────┐
│                    DAEMON                           │
│  Session lifecycle, sandbox pool, trigger engine   │
└────────────────────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ @ayo sandbox│  │Squad sandbox│  │Agent sandbox│
└─────────────┘  └─────────────┘  └─────────────┘
```

**Sandbox providers:**
- **Apple Container** (macOS 26+) – Native Linux containers
- **systemd-nspawn** (Linux) – Lightweight namespace isolation

## Documentation

### Getting Started

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | First 30 minutes with ayo |
| [Core Concepts](docs/concepts.md) | Understanding ayo's model |

### Tutorials

| Tutorial | Description |
|----------|-------------|
| [First Agent](docs/tutorials/first-agent.md) | Create your first custom agent |
| [Working with Squads](docs/tutorials/squads.md) | Multi-agent coordination |
| [Setting up Triggers](docs/tutorials/triggers.md) | Ambient automation |
| [Using Memory](docs/tutorials/memory.md) | Persistent context |
| [Building Plugins](docs/tutorials/plugins.md) | Extending ayo |

### Guides

| Guide | Description |
|-------|-------------|
| [Agents](docs/guides/agents.md) | Agent configuration deep-dive |
| [Squads](docs/guides/squads.md) | Squad management and SQUAD.md |
| [Triggers](docs/guides/triggers.md) | Trigger types and options |
| [Tools](docs/guides/tools.md) | Available tools and custom tools |
| [Sandbox](docs/guides/sandbox.md) | Container configuration |
| [Security](docs/guides/security.md) | Trust levels and guardrails |

### Patterns

| Pattern | Description |
|---------|-------------|
| [Watcher](docs/patterns/watcher.md) | React to file changes |
| [Scheduled](docs/patterns/scheduled.md) | Time-based automation |
| [Ticket Worker](docs/patterns/ticket-worker.md) | Process work queues |
| [Monitor](docs/patterns/monitor.md) | System health checks |

### Reference

| Reference | Description |
|-----------|-------------|
| [CLI Reference](docs/reference/cli.md) | All commands and flags |
| [Configuration](docs/reference/ayo-json.md) | ayo.json schema |
| [Prompts](docs/reference/prompts.md) | Prompt customization |
| [Plugins](docs/reference/plugins.md) | Plugin interface |
| [RPC](docs/reference/rpc.md) | Daemon JSON-RPC API |

### Advanced

| Topic | Description |
|-------|-------------|
| [Architecture](docs/advanced/architecture.md) | System internals |
| [Extending ayo](docs/advanced/extending.md) | Custom tools and providers |
| [Troubleshooting](docs/advanced/troubleshooting.md) | Common issues |

## Troubleshooting

```bash
# Check system health
ayo doctor

# View daemon status
ayo sandbox service status

# Restart daemon
ayo sandbox service stop
ayo sandbox service start

# Shell into a sandbox
ayo sandbox shell @ayo
```

## Requirements

- **macOS 26+** (Tahoe) with Apple Silicon, or **Linux** with systemd
- **Go 1.24+** (for building from source)
- At least one LLM provider configured (Anthropic, OpenAI, Ollama, etc.)


