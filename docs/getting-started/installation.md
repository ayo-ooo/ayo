# Installation

## Prerequisites

- **Go 1.21+**: Required to build generated agents

## Install Ayo CLI

### Homebrew (recommended)

```bash
brew install ayo-ooo/tap/ayo
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

Configure your API key for your preferred provider:

| Environment Variable        | Provider                                           |
| --------------------------- | -------------------------------------------------- |
| `ANTHROPIC_API_KEY`         | Anthropic                                          |
| `OPENAI_API_KEY`            | OpenAI                                             |
| `GEMINI_API_KEY`            | Google Gemini                                      |
| `GROQ_API_KEY`              | Groq                                               |
| `OPENROUTER_API_KEY`        | OpenRouter                                         |
| `CEREBRAS_API_KEY`          | Cerebras                                           |
| `HF_TOKEN`                  | Hugging Face Inference                             |

### Config File

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
