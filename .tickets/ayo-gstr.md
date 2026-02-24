---
id: ayo-gstr
title: Getting Started guide polish
status: closed
priority: high
assignee: "@ayo"
tags: [docs, gtm, polish, first-impression]
created: 2026-02-24
---

# Getting Started guide polish

The Getting Started guide is the primary onboarding document. It must be flawless.

## Current Issues

### 1. Provider Bias (Lines 41-48)

```markdown
# Anthropic (recommended)
export ANTHROPIC_API_KEY="sk-ant-..."
```

**Fix**: Remove "(recommended)", add full provider table, emphasize `ayo setup`.

### 2. Missing Providers

Only shows Anthropic, OpenAI, Ollama. Missing: Google, OpenRouter, Azure, Groq, DeepSeek, Cerebras, xAI, Together.

### 3. Missing Session Continuation

No mention of `-c` or `-s` flags for session continuation.

### 4. Missing `ayo doctor` Details

Line 72-79 shows `ayo doctor` but doesn't explain what each check means or how to fix failures.

## Required Changes

### Configure Section (Lines 36-49)

**Before:**
```markdown
### 1. Configure Your API Key

Set your LLM provider's API key:

# Anthropic (recommended)
export ANTHROPIC_API_KEY="sk-ant-..."
```

**After:**
```markdown
### 1. Configure Your LLM Provider

Run the interactive setup wizard:

\`\`\`bash
ayo setup
\`\`\`

This will detect available providers and guide you through configuration.

**Supported Providers:**

| Provider | Environment Variable |
|----------|---------------------|
| Anthropic | `ANTHROPIC_API_KEY` |
| OpenAI | `OPENAI_API_KEY` |
| Google | `GEMINI_API_KEY` |
| OpenRouter | `OPENROUTER_API_KEY` |
| Ollama | (none - local) |
| [+ 5 more] | See Environment Reference |
```

### Add Session Continuation Section

After "Your First Prompt":

```markdown
## Continuing Sessions

Resume your most recent conversation:

\`\`\`bash
ayo -c "continue where we left off"
\`\`\`

Or continue a specific session:

\`\`\`bash
ayo session list
ayo -s sess_abc123 "let's pick this up"
\`\`\`
```

### Expand ayo doctor Section

```markdown
### 4. Verify Installation

\`\`\`bash
ayo doctor
\`\`\`

Expected output:

\`\`\`
✓ Configuration valid
✓ Daemon running  
✓ Container runtime available
✓ Provider configured (anthropic)
\`\`\`

**Troubleshooting:**

| Issue | Solution |
|-------|----------|
| ✗ Daemon not running | Run `ayo sandbox service start` |
| ✗ No provider configured | Run `ayo setup` |
| ✗ Container runtime unavailable | macOS 26+ required |
```

## Path Locations Section

Add a reference section:

```markdown
## Directory Structure

| Location | Purpose |
|----------|---------|
| `~/.config/ayo/` | Configuration |
| `~/.config/ayo/agents/` | Your agents |
| `~/.config/ayo/ayo.json` | Global config |
| `~/.local/share/ayo/` | Data (sessions, memories) |
| `./.config/ayo/` | Project-local config |
```

## Acceptance Criteria

- [ ] No provider shown as "recommended"
- [ ] Full provider table with all 10+ providers
- [ ] `ayo setup` emphasized as primary method
- [ ] Session continuation documented
- [ ] `ayo doctor` output explained
- [ ] Directory structure documented
- [ ] All commands verified correct

## Dependencies

- ayo-prov (provider neutrality)
- ayo-cmds (command accuracy)
- ayo-path (path consistency)

## Notes

This is the first document most users will read after the README.
