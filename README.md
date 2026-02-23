# ayo – Agents You Orchestrate

**ayo** is a CLI for managing AI agents that work in isolated sandbox environments. Create specialized agents, coordinate them through squads, and set up triggers for ambient automation.

```bash
# Just ask
ayo "help me debug this function"

# Use a specialized agent
ayo @reviewer "review this PR for security issues"

# Dispatch to a team
ayo #dev-team "build a user authentication system"
```

## Quick Start

```bash
# Install
go install github.com/alexcabrera/ayo/cmd/ayo@latest

# Configure
export ANTHROPIC_API_KEY="sk-ant-..."
ayo setup

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
ayo agents create @support

# Use it
ayo @support "help with this customer issue"
```

### Squads (`#team`)

Squads are isolated sandboxes where agents collaborate:

```bash
# Create a team
ayo squad create dev-team --agents @frontend,@backend,@qa

# Dispatch work
ayo #dev-team "build the login feature"
```

The `SQUAD.md` constitution defines roles and coordination rules:

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
ayo ticket create "Implement login API" -a @backend
ayo ticket create "Build login form" -a @frontend --deps <api-ticket>

# See what's ready
ayo ticket ready
```

When a ticket closes, blocked dependents unblock automatically.

### Triggers

Set up ambient agents that act without prompting:

```bash
# Daily standup at 9am
ayo trigger create morning-standup \
  --cron "0 9 * * MON-FRI" \
  --agent @standup \
  --prompt "Summarize yesterday's progress"

# Watch for new files
ayo trigger create code-reviewer \
  --watch ~/Projects/*.go \
  --agent @reviewer \
  --prompt "Review changed files"
```

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

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | First 30 minutes with ayo |
| [Agents](docs/agents.md) | Creating and configuring agents |
| [Squads](docs/squads.md) | Multi-agent team coordination |
| [Triggers](docs/triggers.md) | Ambient and proactive agents |
| [Planners](docs/planners.md) | Todos and ticket-based planning |
| [CLI Reference](docs/cli-reference.md) | All commands |

## Troubleshooting

```bash
# Check system health
ayo doctor

# View daemon status
ayo service status

# Shell into a sandbox
ayo sandbox shell @ayo
```

## License

MIT
