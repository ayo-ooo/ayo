# Installation

## Prerequisites

- **Go 1.21+**: Required to build generated agents
- **API Key**: An API key from a supported provider (Anthropic, OpenAI, etc.)

## Install Ayo CLI

### Homebrew (recommended)

```bash
brew tap ayo-ooo/tap
brew install ayo
```

### Go Install

```bash
go install github.com/ayo-ooo/ayo/cmd/ayo@latest
```

### Verify Installation

```bash
ayo --version
```

## Configure API Keys

Ayo agents use the [Fantasy](https://github.com/charmbracelet/fantasy) library for LLM access. Configure your API keys:

### Option 1: Environment Variables

```bash
export ANTHROPIC_API_KEY=your-key-here
export OPENAI_API_KEY=your-key-here
```

### Option 2: Config File

Create `~/.config/ayo/config.toml`:

```toml
[providers.anthropic]
api_key = "your-key-here"

[providers.openai]
api_key = "your-key-here"
```

## Next Steps

- [Quickstart](quickstart.md) - Build your first agent
- [First Agent](first-agent.md) - Learn agent structure
