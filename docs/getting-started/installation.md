# Installation

## Prerequisites

- **Go 1.21+**: Required to build Ayo and generated agents
- **API Key**: An API key from a supported provider (Anthropic, OpenAI, etc.)

## Install Ayo CLI

### From Source

```bash
git clone https://github.com/charmbracelet/ayo.git
cd ayo
go install ./cmd/ayo
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
