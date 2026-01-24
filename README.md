# `ayo` - Agents You Orchestrate

`ayo` is a command-line tool for running AI agents that can execute tasks, use tools, and chain together via Unix pipes. Define agents with custom system prompts, extend them with skills, and compose them into powerful workflows.

## Getting Started

```bash
go install github.com/alexcabrera/ayo/cmd/ayo@latest
```

That's it. Built-in agents and skills install automatically on first run.

Start chatting with built-in default agent:

```bash
ayo @ayo
```

Shortcut to prompt @ayo single prompt:

```bash
ayo "write a haiku about terminal emulators"
```

Attach files for context:

```bash
ayo -a go.sum "review this code"
```

## Agents

Agents are AIs with custom system prompts and capabilities. Each agent is a directory containing configuration and instructions.

### Built-in Agents

| Agent | Description |
|-------|-------------|
| `@ayo` | The default agent - a versatile command-line assistant |
| `@ayo.agents` | Agent management agent for creating and managing agents |
| `@ayo.research` | Web-enabled research agent for finding information online |
| `@ayo.skills` | Skill management agent for creating and organizing skills |

### Running Agents

```bash
# Interactive chat (continues until Ctrl+C)
ayo @ayo.research

# Single prompt
ayo @ayo.research "what's new in Go 1.22?"
```

### Creating Agents

```bash
ayo agents create @myagent
```

This launches an interactive wizard to configure your agent. For non-interactive creation:

```bash
ayo agents create @helper -n \
  --description "A helpful assistant" \
  --model gpt-4.1 \
  --system "You are concise and friendly."
```

### Agent Structure

```
@myagent/
├── config.json         # Agent configuration
├── system.md           # System prompt
├── skills/             # Agent-specific skills
├── input.jsonschema    # Optional: structured input schema
└── output.jsonschema   # Optional: structured output schema
```

### Managing Agents

```bash
ayo agents list              # List all agents
ayo agents show @ayo         # Show agent details
ayo agents update            # Update built-in agents
```

## Skills

Skills extend agent capabilities with domain-specific instructions. They follow the [agentskills spec](https://agentskills.org).

### Built-in Skills

| Skill | Description |
|-------|-------------|
| `ayo` | CLI documentation for programmatic use |
| `debugging` | Systematic debugging techniques |
| `web-search` | Web search capabilities |

### Creating Skills

```bash
# Create in current directory (project-local)
ayo skills create my-skill

# Create in user shared directory
ayo skills create my-skill --shared
```

### Skill Structure

Each skill is a directory with a `SKILL.md` file:

```
my-skill/
├── SKILL.md            # Required: skill definition with YAML frontmatter
├── scripts/            # Optional: executable code
├── references/         # Optional: additional documentation
└── assets/             # Optional: templates, data files
```

The `SKILL.md` format:

```markdown
---
name: my-skill
description: What this skill does and when to use it.
metadata:
  author: your-name
  version: "1.0"
---

# Skill Instructions

Detailed instructions for the agent...
```

### Managing Skills

```bash
ayo skills list                  # List all skills
ayo skills list --source=built-in # Filter by source
ayo skills show debugging        # Show skill details
ayo skills validate ./my-skill   # Validate a skill directory
ayo skills update                # Update built-in skills
```

## Agent Chaining

Agents with structured I/O schemas can be composed via Unix pipes. The output of one agent becomes the input to the next.

```bash
# Chain agents together
ayo @code-reviewer '{"files":["main.go"]}' | ayo @issue-reporter
```

When stdout is piped:
- UI (spinners, reasoning, tool calls) goes to stderr
- Raw JSON output goes to stdout for downstream consumption

### Chain Discovery

```bash
# List chainable agents
ayo chain ls

# Inspect schemas
ayo chain inspect @myagent

# Find compatible agents
ayo chain from @source-agent    # What can receive this output?
ayo chain to @target-agent      # What can feed this input?

# Validate and test
ayo chain validate @myagent '{"key": "value"}'
ayo chain example @myagent      # Generate example input
```

## Configuration

### Config File

Located at `~/.config/ayo/ayo.json`:

```json
{
  "$schema": "./ayo-schema.json",
  "default_model": "gpt-4.1",
  "provider": {
    "name": "openai"
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
| `GOOGLE_API_KEY` | Google AI API key |

### Directory Structure

ayo uses XDG-style directories:

| Platform | User Config | Built-in Data |
|----------|-------------|---------------|
| macOS/Linux | `~/.config/ayo/` | `~/.local/share/ayo/` |
| Windows | `%LOCALAPPDATA%\ayo\` | `%LOCALAPPDATA%\ayo\` |

```
~/.config/ayo/                    # User configuration
├── ayo.json                      # Main config
├── agents/                       # User agents
├── skills/                       # User skills
└── prompts/                      # Custom system prompts
    ├── prefix.md                 # Prepended to all agents
    └── suffix.md                 # Appended to all agents

~/.local/share/ayo/               # Built-in data
├── agents/                       # Built-in agents
├── skills/                       # Built-in skills
└── .builtin-version              # Version marker
```

### Load Priority

Resources are discovered in this order (first found wins):

1. Agent-specific (in agent's `skills/` directory)
2. Project-local (`./.config/ayo/`)
3. User config (`~/.config/ayo/`)
4. Built-in (`~/.local/share/ayo/`)

This allows project-specific overrides of user and built-in resources.

## Commands

### Root Command

```bash
ayo [command] [@agent] [prompt] [--flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--attachment` | `-a` | File attachments (repeatable) |
| `--config` | | Path to config file |
| `--debug` | | Show raw tool payloads |
| `--help` | `-h` | Help |
| `--version` | `-v` | Version |

### agents

```bash
ayo agents list              # List agents
ayo agents show <handle>     # Show details
ayo agents create <handle>   # Create agent
ayo agents update            # Update built-ins
```

### skills

```bash
ayo skills list              # List skills
ayo skills show <name>       # Show details
ayo skills create <name>     # Create skill
ayo skills validate <path>   # Validate skill
ayo skills update            # Update built-ins
```

### chain

```bash
ayo chain ls                 # List chainable agents
ayo chain inspect <agent>    # Show schemas
ayo chain from <agent>       # Find output consumers
ayo chain to <agent>         # Find input producers
ayo chain validate <agent>   # Validate input JSON
ayo chain example <agent>    # Generate example input
```

### sessions

```bash
ayo sessions list              # List conversation sessions
ayo sessions list -a @ayo      # Filter by agent
ayo sessions show <id>         # Show session details
ayo sessions continue          # Continue a session (interactive picker)
ayo sessions continue <id>     # Continue a specific session
ayo sessions delete <id>       # Delete a session
```

### setup

```bash
ayo setup              # Reinstall built-ins
ayo setup --force      # Overwrite without prompting
```

## Tool System

Agents execute tasks through tools. The default tool is `bash`, which executes shell commands.

When the agent runs a command, you'll see:
- A spinner with the command description
- The command output in a styled box
- Success/failure status with elapsed time

Tools are configured per-agent in `config.json`:

```json
{
  "allowed_tools": ["bash"]
}
```

## Plugins

Plugins extend ayo with additional agents, skills, and tools distributed via git repositories.

### Installing Plugins

```bash
# Install from GitHub
ayo plugins install https://github.com/alexcabrera/ayo-plugins-crush

# Install from any git URL
ayo plugins install https://gitlab.com/org/ayo-plugins-mytools.git
ayo plugins install git@github.com:user/ayo-plugins-custom.git

# Reinstall/update
ayo plugins install https://github.com/user/repo --force
```

### Managing Plugins

```bash
ayo plugins list           # List installed plugins
ayo plugins show <name>    # Show plugin details
ayo plugins update         # Update all plugins
ayo plugins update <name>  # Update specific plugin
ayo plugins remove <name>  # Uninstall plugin
```

### What Plugins Provide

| Component | Description |
|-----------|-------------|
| **Agents** | Custom AI agents with specialized system prompts and tools |
| **Skills** | Domain-specific instructions that extend agent capabilities |
| **Tools** | External CLI commands wrapped as agent tools |
| **Delegates** | Task-type handlers (e.g., coding, research) |

### Example: Crush Coding Plugin

The [crush plugin](https://github.com/alexcabrera/ayo-plugins-crush) adds a powerful coding agent:

```bash
ayo plugins install https://github.com/alexcabrera/ayo-plugins-crush
```

After installation, you can:
- Use `@crush` directly: `ayo @crush "refactor the auth module"`
- Let `@ayo` delegate coding tasks to `@crush` automatically

### Creating Plugins

See [docs/plugins.md](docs/plugins.md) for a complete guide to creating plugins.

## Interactive Mode

In interactive mode, ayo maintains conversation context across turns:

- First `Ctrl+C` interrupts the current request
- Second `Ctrl+C` at the prompt exits the session

```bash
ayo @ayo          # Start interactive session
```

## Piping and Scripting

ayo is designed for Unix pipelines:

```bash
# Pipe input
echo "explain this" | ayo

# Pipe output (UI goes to stderr)
ayo @reporter "analyze logs" | jq .

# Chain agents
ayo @analyzer '{"data":"..."}' | ayo @reporter
```

When stdin or stdout is a pipe, ayo adjusts its behavior:
- UI output always goes to stderr
- Raw content goes to stdout
- JSON output is not markdown-rendered

