# Getting Started with Ayo

Get up and running with ayo in under 5 minutes. By the end of this guide, you'll have run your first AI agent in an isolated sandbox.

## Prerequisites

- **macOS 26+** (Apple Container) or **Linux** (systemd-nspawn)
- **Go 1.22+** (for building from source)
- An LLM API key (Anthropic, OpenAI, or Vertex AI)

## Installation

### From Source (Recommended)

```bash
# Clone and build
git clone https://github.com/alexcabrera/ayo
cd ayo
go install ./cmd/ayo/...

# Verify installation
ayo --help
```

### From Go

```bash
go install github.com/alexcabrera/ayo/cmd/ayo@latest
```

## Setup

### 1. Configure Your API Key

Set your LLM provider's API key:

```bash
# Anthropic (default)
export ANTHROPIC_API_KEY="sk-ant-..."

# OpenAI
export OPENAI_API_KEY="sk-..."

# Google Vertex AI
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
```

### 2. Run Setup Wizard

```bash
ayo setup
```

This will:
- Configure your LLM provider
- Create the default `@ayo` sandbox
- Start the background daemon

### 3. Verify Installation

```bash
ayo doctor
```

This checks that all components are working correctly.

## Your First Prompt

Talk to the default `@ayo` agent:

```bash
ayo "Hello! What can you help me with?"
```

The agent runs inside an isolated sandbox and can:
- Execute shell commands
- Read and write files (with your approval)
- Remember context across sessions

## Interactive Mode

Start an interactive session:

```bash
ayo
```

Type your prompts, and press Enter to send. Use `Ctrl+C` or type `exit` to quit.

## Attach Files

Include files in your prompt:

```bash
ayo -a main.go "explain this code"
ayo -a src/ "find bugs in this directory"
```

## Continue Sessions

Resume your last conversation:

```bash
ayo -c "also add error handling"
```

Or continue a specific session:

```bash
ayo -s abc123 "one more thing..."
```

## Create a Custom Agent

Create an agent specialized for a task:

```bash
ayo agent create @reviewer
```

This creates a directory at `~/.config/ayo/agents/@reviewer/` with:
- `config.json` – Model, tools, and settings
- `system.md` – The agent's behavior instructions

Edit `system.md` to customize behavior:

```markdown
You are a code review specialist. Focus on:
- Security vulnerabilities
- Performance issues
- Code style consistency

Be direct and specific. Cite line numbers.
```

Use your agent:

```bash
ayo @reviewer "review the changes in src/"
```

## Create a Squad

Squads are teams of agents that collaborate:

```bash
# Create a squad
ayo squad create dev-team

# Add agents
ayo squad add-agent dev-team @frontend
ayo squad add-agent dev-team @backend

# Dispatch a task
ayo #dev-team "build the user authentication feature"
```

## Set Up a Trigger

Create an agent that runs automatically:

```bash
# Run every morning at 9am
ayo trigger create standup \
  --cron "0 9 * * MON-FRI" \
  --agent @standup \
  --prompt "Summarize yesterday's git commits"

# Watch for file changes
ayo trigger create watcher \
  --watch ./src \
  --agent @linter \
  --prompt "Check changed files for issues"
```

## Understanding Sandbox Isolation

When an agent wants to modify files on your system, you'll see an approval prompt:

```
┌─────────────────────────────────────────────┐
│ @ayo wants to write:                        │
│   ~/Projects/app/main.go                    │
│                                             │
│ [Y]es  [N]o  [D]iff  [A]lways for session   │
└─────────────────────────────────────────────┘
```

- **Y** – Approve this change
- **N** – Deny and tell the agent
- **D** – View the diff first
- **A** – Auto-approve similar requests this session

To auto-approve all changes (use with caution):

```bash
ayo --no-jodas "refactor this file"
```

## Key Commands Reference

| Command | Description |
|---------|-------------|
| `ayo [prompt]` | Chat with default @ayo agent |
| `ayo @agent [prompt]` | Chat with specific agent |
| `ayo #squad [prompt]` | Dispatch to squad |
| `ayo agent list` | List available agents |
| `ayo agent create @name` | Create new agent |
| `ayo squad list` | List squads |
| `ayo squad create name` | Create new squad |
| `ayo trigger list` | List active triggers |
| `ayo memory list` | Show stored memories |
| `ayo doctor` | Check system health |
| `ayo service status` | Check daemon status |

## Next Steps

Now that you're up and running:

- [Core Concepts](concepts.md) – Understand the mental model
- [Creating Agents](tutorials/first-agent.md) – Build a specialized agent
- [Multi-Agent Squads](tutorials/squads.md) – Set up team coordination
- [Triggers & Automation](tutorials/triggers.md) – Create ambient agents
- [Memory System](tutorials/memory.md) – Persistent context

## Troubleshooting

### Daemon won't start

```bash
# Check status
ayo service status

# Remove stale socket and restart
rm -f ~/.local/share/ayo/daemon.sock
ayo service start
```

### Sandbox creation fails

```bash
# Check system requirements
ayo doctor

# On macOS: Ensure macOS 26+ for Apple Container
# On Linux: Ensure systemd-nspawn is installed
```

### Agent not found

```bash
# List available agents
ayo agent list

# Check agent directories
ls ~/.config/ayo/agents/
ls ~/.local/share/ayo/agents/
```

---

*Ready to dive deeper? Continue to [Core Concepts](concepts.md).*
