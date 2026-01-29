# Getting Started

This guide will help you install ayo and start using AI agents in your terminal.

## Installation

### From Source (Go)

```bash
go install github.com/alexcabrera/ayo/cmd/ayo@latest
```

### From Binary

Download the latest release from [GitHub Releases](https://github.com/alexcabrera/ayo/releases).

### Verify Installation

```bash
ayo --version
```

## First Run

Built-in agents and skills install automatically on first run:

```bash
ayo
```

This will:
1. Install built-in agents to `~/.local/share/ayo/agents/`
2. Install built-in skills to `~/.local/share/ayo/skills/`
3. Create config directory at `~/.config/ayo/`
4. Start an interactive chat with `@ayo`

## Prerequisites

### API Keys

Ayo requires an API key from at least one LLM provider:

| Provider | Environment Variable |
|----------|---------------------|
| OpenAI | `OPENAI_API_KEY` |
| Anthropic | `ANTHROPIC_API_KEY` |
| OpenRouter | `OPENROUTER_API_KEY` |
| Google AI | `GOOGLE_API_KEY` |

Set your API key:

```bash
export OPENAI_API_KEY="sk-..."
```

### Ollama (Optional)

For memory features, ayo uses Ollama to run small local models:

```bash
# Install Ollama (macOS)
brew install ollama

# Start Ollama service
ollama serve

# Pull required models
ollama pull nomic-embed-text
ollama pull ministral-3:3b
```

Check everything is working:

```bash
ayo doctor
```

## Basic Usage

### Interactive Chat

Start a conversation with the default agent:

```bash
ayo
```

Exit with `Ctrl+C` (twice if mid-response).

### Single Prompt

Run a prompt and exit:

```bash
ayo "list all files in the current directory"
```

### File Attachments

Attach files for context:

```bash
ayo -a main.go "review this code"
ayo -a file1.txt -a file2.txt "compare these files"
```

## The @ayo Agent

`@ayo` is the default and only built-in agent. It's a versatile assistant that can:

- Execute shell commands via the `bash` tool
- Delegate to other agents via `agent_call`
- Track multi-step tasks with the `todo` tool
- Create and manage other agents and skills
- Use any attached skills for specialized tasks

```bash
# List all available agents
ayo agents list

# Ask @ayo to help with anything
ayo "help me create an agent for code review"
ayo "what skills are available?"
ayo "debug this test failure"
```

## Architecture Overview

```
~/.config/ayo/           # User configuration
├── ayo.json             # Main config (model, provider)
├── agents/              # Your custom agents
└── skills/              # Your custom skills

~/.local/share/ayo/      # Data and built-ins
├── ayo.db               # Sessions and memories
├── agents/              # Built-in agents (@ayo)
├── skills/              # Built-in skills
└── plugins/             # Installed plugins
```

## Common Workflows

### Create a Custom Agent

```bash
# Ask @ayo to help design it
ayo "help me create an agent for debugging Go code"

# Or use the CLI directly
ayo agents create @debugger \
  -m gpt-5.2 \
  -d "Debugging specialist" \
  -f ~/prompts/debugger.md
```

### Install a Plugin

```bash
# Add research capabilities
ayo plugins install https://github.com/user/ayo-plugins-research

# Use the new agent
ayo @research "latest developments in AI"
```

### Resume a Conversation

```bash
# List recent sessions
ayo sessions list

# Continue a session
ayo sessions continue
```

## Next Steps

- [Create your first agent](agents.md#creating-agents)
- [Learn about skills](skills.md)
- [Set up memory](memory.md)
- [Install plugins](plugins.md) for additional capabilities
- [CLI Reference](cli-reference.md) for all commands

## Troubleshooting

### Check System Health

```bash
ayo doctor
ayo doctor -v  # Verbose output with model list
```

### Common Issues

**"No API key found"**
- Set `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, or another provider's key

**"Model not found"**
- Check `ayo doctor -v` for available models
- Update `default_model` in `~/.config/ayo/ayo.json`

**"Ollama not running"**
- Start with `ollama serve`
- Memory features require Ollama (optional)

**Agent not found**
- Run `ayo setup` to reinstall built-ins
- Check `ayo agents list` for available agents
