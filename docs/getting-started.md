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
4. Start an interactive chat with the default `@ayo` agent

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

Or specify an agent:

```bash
ayo @ayo
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

## Built-in Agents

| Agent | Description |
|-------|-------------|
| `@ayo` | Default versatile assistant |
| `@ayo.agents` | Agent management (creating/modifying agents) |
| `@ayo.skills` | Skill management (creating/modifying skills) |

```bash
# List all available agents
ayo agents list
```

## Next Steps

- [Create your first agent](agents.md#creating-agents)
- [Learn about skills](skills.md)
- [Set up memory](memory.md)
- [Install plugins](plugins.md) for additional capabilities

## Troubleshooting

### Check System Health

```bash
ayo doctor
ayo doctor -v  # Verbose output
```

### Common Issues

**"No API key found"**
- Set `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, or another provider's key

**"Model not found"**
- Check `ayo doctor -v` for available models
- Update `default_model` in `~/.config/ayo/ayo.json`

**"Ollama not running"**
- Start with `ollama serve`
- Memory features require Ollama

**Agent not found**
- Run `ayo setup` to reinstall built-ins
- Check `ayo agents list` for available agents
