---
id: ayo-env
title: Complete environment variable documentation
status: closed
priority: high
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# Complete environment variable documentation

Create a comprehensive environment variable reference.

## Problem

`docs/reference/cli.md:588-596` only lists 2 environment variables but many more are supported.

## Complete Environment Variable List

### LLM Provider API Keys

| Variable | Provider | Notes |
|----------|----------|-------|
| `ANTHROPIC_API_KEY` | Anthropic | Claude models |
| `OPENAI_API_KEY` | OpenAI | GPT models |
| `GEMINI_API_KEY` | Google | Gemini models |
| `OPENROUTER_API_KEY` | OpenRouter | Multi-provider gateway |
| `AZURE_OPENAI_API_KEY` | Azure | Azure OpenAI service |
| `GROQ_API_KEY` | Groq | Fast inference |
| `DEEPSEEK_API_KEY` | DeepSeek | DeepSeek models |
| `CEREBRAS_API_KEY` | Cerebras | Fast inference |
| `XAI_API_KEY` | xAI | Grok models |
| `TOGETHER_API_KEY` | Together.ai | Open models |

Source: `internal/config/credentials.go:24-35`

### Service Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama endpoint for local models |
| `CATWALK_URL` | `http://localhost:8080` | Catwalk service URL |

### Ayo Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AYO_CONFIG` | `~/.config/ayo/ayo.json` | Config file path |
| `AYO_LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `XDG_CONFIG_HOME` | `~/.config` | Base config directory |
| `XDG_DATA_HOME` | `~/.local/share` | Base data directory |

### Embedding Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AYO_EMBEDDING_PROVIDER` | `ollama` | Embedding provider |
| `AYO_EMBEDDING_MODEL` | `nomic-embed-text` | Embedding model |

## Where to Document

### Option 1: Expand cli.md Environment section

Add complete table to `docs/reference/cli.md:588-596`

### Option 2: Create separate reference page

Create `docs/reference/environment.md` with full documentation

## Acceptance Criteria

- [ ] All 10+ provider API keys documented
- [ ] Service URLs documented (OLLAMA_HOST, CATWALK_URL)
- [ ] Ayo config variables documented
- [ ] Embedding variables documented
- [ ] XDG variables explained

## Dependencies

- ayo-prov (provider neutrality - ensure table is neutral)

## Notes

Users should be able to find any environment variable in one location.
