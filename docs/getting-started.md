# Getting Started with Ayo

Get up and running with ayo in under 5 minutes. By the end of this guide, you'll have run your first AI agent in an isolated sandbox.

## Prerequisites

- **macOS 26+** (Apple Container) or **Linux** (systemd-nspawn)
- **Go 1.24+** (for building from source)
- An LLM API key (Anthropic, OpenAI, Ollama, or Vertex AI)

## Installation

### From Source (Recommended)

```bash
# Clone and build
git clone https://github.com/alexcabrera/ayo
cd ayo
go build -o ayo ./cmd/ayo/...

# Add to PATH or move to a bin directory
mv ayo /usr/local/bin/

# Verify installation
ayo --help
```

### From Go

```bash
go install github.com/alexcabrera/ayo/cmd/ayo@latest
```

## Setup

### 1. Configure Your LLM Provider

Run the interactive setup wizard:

```bash
ayo setup
```

This detects available API keys and guides you through configuration.

**Supported Providers:**

| Provider | Environment Variable | Notes |
|----------|---------------------|-------|
| Anthropic | `ANTHROPIC_API_KEY` | Claude models |
| OpenAI | `OPENAI_API_KEY` | GPT-4 models |
| Google | `GEMINI_API_KEY` | Gemini models |
| OpenRouter | `OPENROUTER_API_KEY` | Multi-provider gateway |
| Azure | `AZURE_OPENAI_API_KEY` | Azure OpenAI |
| Groq | `GROQ_API_KEY` | Fast inference |
| DeepSeek | `DEEPSEEK_API_KEY` | DeepSeek models |
| Cerebras | `CEREBRAS_API_KEY` | Fast inference |
| xAI | `XAI_API_KEY` | Grok models |
| Together | `TOGETHER_API_KEY` | Open models |
| Ollama | *(none required)* | Local models |

Set your preferred provider's API key before running setup:

```bash
export YOUR_PROVIDER_API_KEY="your-key-here"
ayo setup
```

### 2. Run Setup Wizard

```bash
ayo setup
```

This will:
- Create configuration directory (`~/.config/ayo/`)
- Create data directories (`~/.local/share/ayo/`)
- Install default agents
- Configure your LLM provider

### 3. Start the Daemon

```bash
ayo service start
```

### 4. Verify Installation

```bash
ayo doctor
```

This checks that all components are working correctly:
- ✓ Configuration valid
- ✓ Daemon running
- ✓ Container runtime available
- ✓ Provider configured

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

Type your prompts and press Enter to send. Use `/exit` or `Ctrl+D` to quit.

## Attach Files

Include files in your prompt:

```bash
ayo "explain this code" -a main.go
ayo "find bugs in this directory" -a src/
```

## Continue Sessions

Resume your last conversation:

```bash
ayo -c "also add error handling"
```

Or continue a specific session:

```bash
ayo -s <session-id> "one more thing..."
```

## Create a Custom Agent

Create an agent specialized for a task:

```bash
ayo agent create @reviewer \
  --description "Code review specialist" \
  --prompt "You are a code review expert. Focus on security, performance, and style."
```

Use your agent:

```bash
ayo @reviewer "review the changes in src/"
```

For more complex agents, edit the files directly:
- `~/.local/share/ayo/agents/reviewer/config.json` – Settings
- `~/.local/share/ayo/agents/reviewer/system.md` – Behavior

## Create a Squad

Squads are teams of agents that collaborate:

```bash
# Create a squad with agents
ayo squad create dev-team -a @frontend,@backend

# Start the squad
ayo squad start dev-team

# Dispatch a task
ayo "#dev-team" "build the user authentication feature"
```

## Set Up a Trigger

Create an agent that runs automatically:

```bash
# Run every morning at 9am (weekdays)
ayo trigger schedule @reporter "0 9 * * 1-5" \
  --prompt "Summarize yesterday's git commits"

# Watch for file changes
ayo trigger watch ~/Code/project @linter \
  --prompt "Check changed files for issues" \
  --pattern "*.go"
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

## Key Commands Reference

| Command | Description |
|---------|-------------|
| `ayo "prompt"` | Chat with default @ayo agent |
| `ayo @agent "prompt"` | Chat with specific agent |
| `ayo "#squad" "prompt"` | Dispatch to squad |
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
ls ~/.local/share/ayo/agents/
```

---

*Ready to dive deeper? Continue to [Core Concepts](concepts.md).*
