# Configuration

Ayo uses a layered configuration system with directories for user config, built-in data, and project-local overrides.

## Directory Structure

### Unix (macOS, Linux)

| Directory | Purpose |
|-----------|---------|
| `~/.config/ayo/` | User configuration (editable) |
| `~/.local/share/ayo/` | Built-in data (auto-installed) |

### Windows

Both stored in `%LOCALAPPDATA%\ayo\`

### Full Layout

```
~/.config/ayo/                    # User configuration
├── ayo.json                      # Main config file
├── ayo-schema.json               # JSON schema for config (auto-installed)
├── agents/                       # User-defined agents
│   └── @myagent/
│       ├── config.json
│       ├── system.md
│       └── skills/               # Agent-specific skills
├── skills/                       # User-defined shared skills
│   └── my-skill/
│       └── SKILL.md
└── prompts/                      # Custom system prompts
    ├── system-prefix.md          # Prepended to all agents
    └── system-suffix.md          # Appended to all agents

~/.local/share/ayo/               # Built-in data
├── agents/                       # Built-in agents
│   └── @ayo/
├── skills/                       # Built-in shared skills
│   └── debugging/
├── plugins/                      # Installed plugins
│   └── research/
├── ayo.db                        # SQLite database (sessions, memories)
├── packages.json                 # Plugin registry
└── .builtin-version              # Version marker
```

## Configuration File

Located at `~/.config/ayo/ayo.json`:

```json
{
  "$schema": "./ayo-schema.json",
  "default_model": "gpt-4.1",
  "provider": {
    "name": "openai"
  },
  "delegates": {
    "coding": "@crush",
    "research": "@research"
  },
  "default_tools": {
    "search": "searxng"
  }
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `$schema` | string | Path to JSON schema for editor support |
| `default_model` | string | Default model for agents without explicit model |
| `provider` | object | Provider configuration (see below) |
| `delegates` | object | Task type to agent mappings |
| `default_tools` | object | Tool aliases (e.g., `search` → `searxng`) |
| `agents_dir` | string | Override user agents directory |
| `skills_dir` | string | Override user skills directory |
| `system_prefix` | string | Path to prefix prompt file |
| `system_suffix` | string | Path to suffix prompt file |

### Provider Configuration

```json
{
  "provider": {
    "name": "openai"
  }
}
```

Supported providers:
- `openai` - OpenAI API
- `anthropic` - Anthropic API
- `google` - Google AI API
- `openrouter` - OpenRouter (multiple providers)

## Environment Variables

### API Keys

| Variable | Provider |
|----------|----------|
| `OPENAI_API_KEY` | OpenAI |
| `ANTHROPIC_API_KEY` | Anthropic |
| `OPENROUTER_API_KEY` | OpenRouter |
| `GOOGLE_API_KEY` | Google AI |

### Ollama

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |

## Load Priority

Resources are discovered in this order (first found wins):

1. **Agent-specific** - Skills in agent's `skills/` directory
2. **Project-local** - `./.config/ayo/` in current directory
3. **User config** - `~/.config/ayo/`
4. **Built-in** - `~/.local/share/ayo/`

This allows project-specific overrides of user and built-in resources.

## Project Configuration

Create `.ayo.json` in your project root to configure ayo for that directory:

```json
{
  "agent": "@ayo",
  "model": "gpt-4.1",
  "delegates": {
    "coding": "@crush"
  }
}
```

### Fields

| Field | Description |
|-------|-------------|
| `agent` | Default agent for this directory |
| `model` | Override default model |
| `delegates` | Task type mappings (overrides global) |

Ayo searches from the current directory up to find `.ayo.json`.

## Dev Mode

When running from a source checkout, ayo uses project-local directories:

```
./ayo-main/                       # Source checkout
├── .config/ayo/                  # Project-local user config
│   ├── agents/
│   └── skills/
└── .local/share/ayo/             # Project-local built-in data
    ├── agents/
    └── skills/
```

This allows:
- Testing without affecting user config
- Multiple dev branches with isolated data
- Development plugins in local directories

Enable with the `--dev` flag on setup commands, or detected automatically when running from a git checkout that contains the ayo source.

## Custom Prompts

Add custom prefix/suffix prompts that layer on top of guardrails:

### Prefix Prompt

`~/.config/ayo/prompts/system-prefix.md`:

```markdown
## Project Context

This is the myproject repository. Key conventions:
- Use TypeScript for all new code
- Follow the existing code style
- Always add tests for new features
```

### Suffix Prompt

`~/.config/ayo/prompts/system-suffix.md`:

```markdown
## Additional Guidelines

- Be concise
- Prefer simple solutions
```

## Verifying Configuration

```bash
# Check system health and configuration
ayo doctor

# Verbose output with model list
ayo doctor -v
```
