# CLI Reference

Complete command-line reference for ayo. Each command is documented with its full syntax, options, and example inputs/outputs.

## Command Index

| Command | Description |
|---------|-------------|
| [ayo](cli-ayo.md) | Root command and chat interface |
| [ayo agents](cli-agents.md) | Manage AI agents |
| [ayo ticket](cli-ticket.md) | Task coordination system |
| [ayo squad](cli-squad.md) | Multi-agent team sandboxes |
| [ayo flow](cli-flow.md) | Composable workflows |
| [ayo sandbox](cli-sandbox.md) | Container management |
| [ayo session](cli-session.md) | Conversation persistence |
| [ayo memory](cli-memory.md) | Semantic memory |
| [ayo skill](cli-skill.md) | Reusable instruction modules |
| [ayo share](cli-share.md) | Host directory sharing |
| [ayo trigger](cli-trigger.md) | Scheduled and file-triggered automation |
| [ayo plugin](cli-plugin.md) | Extension management |

## Global Flags

These flags work with all commands:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--config` | | string | Path to config file |
| `--json` | | bool | Output in JSON format |
| `--quiet` | `-q` | bool | Suppress informational messages |
| `--help` | `-h` | bool | Show help |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Configuration error |
| 4 | Network/API error |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `AYO_CONFIG` | Path to config file |
| `AYO_HOME` | Data directory (default: `~/.local/share/ayo`) |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENAI_API_KEY` | OpenAI API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
