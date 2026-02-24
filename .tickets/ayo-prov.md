---
id: ayo-prov
title: Provider-neutral documentation
status: closed
priority: critical
assignee: "@ayo"
tags: [docs, gtm, polish]
created: 2026-02-24
---

# Provider-neutral documentation

Remove all provider bias from documentation. ayo supports multiple LLM providers and docs should not favor any specific one.

## Problem

Documentation currently shows Anthropic as "(recommended)" and uses hardcoded Claude model names throughout. This creates a poor first impression for users of other providers.

## Files requiring changes

### High Priority - First Impression

| File | Line(s) | Issue |
|------|---------|-------|
| `README.md` | 23 | `export ANTHROPIC_API_KEY` shown first/only |
| `docs/getting-started.md` | 41-42 | "Anthropic (recommended)" label |
| `docs/getting-started.md` | 42-48 | Only shows Anthropic/OpenAI/Ollama |

### Model Name Hardcoding

| File | Line(s) | Hardcoded Model |
|------|---------|-----------------|
| `docs/guides/agents.md` | 21, 66, 330 | `claude-sonnet-4-20250514` |
| `docs/guides/squads.md` | 112-117 | `claude-sonnet-4-20250514` |
| `docs/tutorials/first-agent.md` | 94, 209 | `claude-sonnet-4-20250514` |
| `docs/tutorials/memory.md` | 271 | `claude-sonnet-4-20250514` |
| `docs/tutorials/plugins.md` | 77 | `claude-sonnet-4-20250514` |
| `docs/reference/ayo-json.md` | 28, 53, 69 | `claude-sonnet-4-20250514` |
| `docs/_E2E_MANUAL_HUMAN_TESTING_GUIDE.md` | 144, 160, 311 | `claude-sonnet-4-20250514` |
| `docs/advanced/troubleshooting.md` | 617 | `claude-3-haiku-20240307` |

## Required Changes

### 1. Create provider configuration table

Add to getting-started.md and README.md:

```markdown
## Supported Providers

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
| Ollama | (none) | Local models |
```

### 2. Replace hardcoded models with placeholders

Change:
```json
"model": "claude-sonnet-4-20250514"
```

To:
```json
"model": "your-model-here"  // e.g., claude-sonnet-4-20250514, gpt-4o, gemini-pro
```

### 3. Update Quick Start in README.md

Change:
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

To:
```bash
# Set your API key (see Supported Providers below)
ayo setup  # Interactive setup wizard
```

### 4. Remove "(recommended)" from Anthropic

The setup wizard can recommend based on what's available, but docs should be neutral.

## Acceptance Criteria

- [ ] No hardcoded model names in examples (use placeholders with comment examples)
- [ ] No provider shown as "recommended" 
- [ ] Complete provider table with all 10+ supported providers
- [ ] `ayo setup` emphasized as the primary configuration method
- [ ] All environment variables documented in one canonical location

## Dependencies

None

## Notes

Provider list from `internal/config/credentials.go:24-35`
