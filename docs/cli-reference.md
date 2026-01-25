# CLI Reference

Complete command reference for ayo.

## Root Command

```bash
ayo [command] [@agent] [prompt] [--flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--attachment` | `-a` | File attachments (repeatable) |
| `--config` | | Path to config file |
| `--debug` | | Show debug output including raw tool payloads |
| `--help` | `-h` | Help for ayo |
| `--version` | `-v` | Show version |

### Examples

```bash
# Interactive chat with default agent
ayo

# Interactive chat with specific agent
ayo @ayo

# Single prompt
ayo "explain this error"

# With file attachment
ayo -a main.go "review this code"

# Multiple attachments
ayo -a file1.txt -a file2.txt "compare these"
```

---

## ayo agents

Manage agents.

### ayo agents list

List all available agents.

```bash
ayo agents list
```

### ayo agents show

Show agent details.

```bash
ayo agents show <handle>
```

### ayo agents create

Create a new agent.

```bash
ayo agents create @handle [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--model` | `-m` | Model to use |
| `--description` | `-d` | Brief description |
| `--system` | `-s` | System prompt text |
| `--system-file` | `-f` | Path to system prompt file |
| `--tools` | `-t` | Allowed tools (comma-separated) |
| `--skills` | | Skills to include |
| `--exclude-skills` | | Skills to exclude |
| `--ignore-builtin-skills` | | Don't load built-in skills |
| `--ignore-shared-skills` | | Don't load user shared skills |
| `--input-schema` | | JSON schema for stdin input |
| `--output-schema` | | JSON schema for stdout output |
| `--no-guardrails` | | Disable safety guardrails |

**Examples:**

```bash
# Show help (no arguments)
ayo agents create

# Minimal agent
ayo agents create @helper -m gpt-4.1

# With system file
ayo agents create @reviewer -m gpt-4.1 -f system.md

# Full options
ayo agents create @debugger \
  -m gpt-4.1 \
  -d "Debugging specialist" \
  -f system.md \
  -t bash,agent_call \
  --skills debugging
```

**Conversational alternative:**

```bash
ayo "help me create an agent for code review"
```

### ayo agents update

Update built-in agents.

```bash
ayo agents update [--force]
```

---

## ayo skills

Manage skills.

### ayo skills list

List all available skills.

```bash
ayo skills list [--source=<source>]
```

| Flag | Description |
|------|-------------|
| `--source` | Filter by source: `built-in`, `user`, `plugin` |

### ayo skills show

Show skill details.

```bash
ayo skills show <name>
```

### ayo skills create

Create a new skill.

```bash
ayo skills create <name> [--flags]
```

| Flag | Description |
|------|-------------|
| `--shared` | Create in shared skills directory |

### ayo skills validate

Validate a skill directory.

```bash
ayo skills validate <path>
```

### ayo skills update

Update built-in skills.

```bash
ayo skills update [--force]
```

---

## ayo sessions

Manage conversation sessions.

### ayo sessions list

List conversation sessions.

```bash
ayo sessions list [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--agent` | `-a` | Filter by agent handle |
| `--source` | | Filter by source |
| `--limit` | `-l` | Maximum results |

### ayo sessions show

Show session details and conversation.

```bash
ayo sessions show <session-id>
```

### ayo sessions continue

Continue a previous session.

```bash
ayo sessions continue [session-id]
```

Without ID, shows interactive picker.

| Flag | Description |
|------|-------------|
| `--debug` | Show debug output |

### ayo sessions delete

Delete a session.

```bash
ayo sessions delete <session-id> [--force]
```

---

## ayo memory

Manage agent memories.

### ayo memory list

List memories.

```bash
ayo memory list [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--agent` | `-a` | Filter by agent |
| `--category` | `-c` | Filter by category |
| `--limit` | `-l` | Maximum results |
| `--json` | | JSON output |

### ayo memory search

Search memories semantically.

```bash
ayo memory search <query> [--flags]
```

| Flag | Description |
|------|-------------|
| `--agent` | Filter by agent |
| `--threshold` | Similarity threshold (0-1) |
| `--limit` | Maximum results |

### ayo memory show

Show memory details.

```bash
ayo memory show <id>
```

### ayo memory store

Store a new memory.

```bash
ayo memory store <content> [--flags]
```

| Flag | Description |
|------|-------------|
| `--category` | Category: preference, fact, correction, pattern |

### ayo memory forget

Forget a memory (soft delete).

```bash
ayo memory forget <id> [--force]
```

### ayo memory stats

Show memory statistics.

```bash
ayo memory stats
```

### ayo memory clear

Clear all memories.

```bash
ayo memory clear [--flags]
```

| Flag | Description |
|------|-------------|
| `--agent` | Clear for specific agent only |
| `--force` | Skip confirmation |

---

## ayo plugins

Manage plugins.

### ayo plugins install

Install a plugin from git.

```bash
ayo plugins install <git-url> [--flags]
```

| Flag | Description |
|------|-------------|
| `--force` | Reinstall/overwrite |
| `--local` | Install from local directory |
| `--yes` | Skip prompts |

**Examples:**

```bash
# From GitHub
ayo plugins install https://github.com/user/ayo-plugins-name

# Local development
ayo plugins install --local ./my-plugin
```

### ayo plugins list

List installed plugins.

```bash
ayo plugins list
```

### ayo plugins show

Show plugin details.

```bash
ayo plugins show <name>
```

### ayo plugins update

Update plugins.

```bash
ayo plugins update [name] [--flags]
```

| Flag | Description |
|------|-------------|
| `--force` | Force update |
| `--dry-run` | Check without applying |

### ayo plugins remove

Remove a plugin.

```bash
ayo plugins remove <name> [--yes]
```

---

## ayo chain

Explore and validate agent chaining.

### ayo chain ls

List chainable agents.

```bash
ayo chain ls [--json]
```

### ayo chain inspect

Show agent schemas.

```bash
ayo chain inspect <agent> [--json]
```

### ayo chain from

Find agents that can receive this agent's output.

```bash
ayo chain from <agent>
```

### ayo chain to

Find agents whose output this agent can receive.

```bash
ayo chain to <agent>
```

### ayo chain validate

Validate JSON against agent's input schema.

```bash
ayo chain validate <agent> [json]
```

JSON can be provided as argument or via stdin.

### ayo chain example

Generate example input for an agent.

```bash
ayo chain example <agent>
```

---

## ayo setup

Complete ayo setup.

```bash
ayo setup [--flags]
```

| Flag | Description |
|------|-------------|
| `--force` | Overwrite modifications |

This command:
- Installs built-in agents and skills
- Creates user directories
- Checks Ollama models

---

## ayo doctor

Check system health and dependencies.

```bash
ayo doctor [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--verbose` | `-v` | Show detailed output including model list |

Checks:
- Ayo version
- Config file
- Database connection
- Ollama service and models
- Default model configuration

---

## ayo completion

Generate shell completion scripts.

```bash
ayo completion <shell>
```

Supported shells: `bash`, `zsh`, `fish`, `powershell`

### Setup

**Bash:**
```bash
ayo completion bash > /etc/bash_completion.d/ayo
```

**Zsh:**
```bash
ayo completion zsh > "${fpath[1]}/_ayo"
```

**Fish:**
```bash
ayo completion fish > ~/.config/fish/completions/ayo.fish
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
| `GOOGLE_API_KEY` | Google AI API key |
| `OLLAMA_HOST` | Ollama server URL (default: localhost:11434) |

---

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 130 | Interrupted (Ctrl+C) |
