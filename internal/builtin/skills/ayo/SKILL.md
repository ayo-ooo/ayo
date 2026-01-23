---
name: ayo
description: Manage ayo agents, skills, and configuration using the ayo CLI. Use when the user wants to create, list, or modify agents and skills.
metadata:
  author: ayo
  version: "1.0"
---

# Ayo CLI Skill

This skill provides instructions for using the ayo command-line interface to manage agents, skills, and configuration.

## When to Use

Activate this skill when:
- User wants to create a new agent
- User wants to list or inspect existing agents
- User wants to create or manage skills
- User wants to continue or manage conversation sessions
- User asks about agent chaining or structured I/O
- User wants to configure ayo

## CLI Overview

```
ayo [command] [@agent] [prompt] [--flags]
```

### Top-Level Commands

| Command | Description |
|---------|-------------|
| `ayo @agent "prompt"` | Run a prompt with the specified agent |
| `ayo agents` | Manage agents (list, create, show, update) |
| `ayo skills` | Manage skills (list, create, show, validate, update) |
| `ayo sessions` | Manage conversation sessions |
| `ayo chain` | Explore and validate agent chaining |
| `ayo setup` | Install/update built-in agents and skills |

## Running Agents

```bash
# Interactive chat with default agent
ayo

# Interactive chat with specific agent
ayo @agent-name

# Non-interactive: run single prompt and exit
ayo @agent-name "Your prompt here"

# With file attachment
ayo @agent-name -a file.txt "Analyze this file"
```

### Built-in Agents

| Agent | Description |
|-------|-------------|
| `@ayo` | Default versatile assistant with bash and agent delegation |
| `@ayo.crush` | Coding agent powered by Crush for complex source code tasks |
| `@ayo.research` | Research assistant with web search capabilities |
| `@ayo.agents` | Agent management for creating/modifying agents |
| `@ayo.skills` | Skill management for creating/modifying skills |

### Using @ayo.crush

For complex coding tasks, use the `@ayo.crush` agent:

```bash
# Direct invocation
ayo @ayo.crush "Refactor the authentication module to use JWT tokens"

# Multi-file changes
ayo @ayo.crush "Add comprehensive error handling to all database operations in internal/db/"

# With context
ayo @ayo.crush "Fix the failing tests in internal/session/ and ensure all edge cases are covered"
```

@ayo.crush requires Crush to be installed: `go install github.com/charmbracelet/crush@latest`

## Agent Management

### List Agents

```bash
ayo agents list
```

Shows all available agents grouped by source (user-defined vs built-in).

### Show Agent Details

```bash
ayo agents show @agent-name
```

Displays agent configuration including model, tools, skills, and location.

### Create Agent

Interactive wizard:
```bash
ayo agents create @my-agent
```

Non-interactive (skip wizard):
```bash
ayo agents create @my-agent \
  --non-interactive \
  --model gpt-4.1 \
  --description "My agent description" \
  --system "You are a helpful assistant..." \
  --tools bash,agent_call \
  --skills debugging
```

#### Create Agent Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--model` | `-m` | Model to use (required in non-interactive mode) |
| `--description` | `-d` | Brief description of the agent |
| `--system` | `-s` | System prompt text |
| `--system-file` | `-f` | Path to system prompt file (.md or .txt) |
| `--tools` | `-t` | Allowed tools: bash, agent_call (comma-separated) |
| `--skills` | | Skills to include (comma-separated) |
| `--exclude-skills` | | Skills to exclude |
| `--ignore-builtin-skills` | | Don't load built-in skills |
| `--ignore-shared-skills` | | Don't load user shared skills |
| `--input-schema` | | JSON schema for validating stdin input |
| `--output-schema` | | JSON schema for structuring stdout output |
| `--no-system-wrapper` | | Disable system guardrails (not recommended) |
| `--non-interactive` | `-n` | Skip wizard, fail if required fields missing |
| `--dev` | | Create in local ./.config/ayo/ directory |

#### Using System Prompt Files

Create a markdown file with the system prompt:
```bash
cat > system.md << 'EOF'
You are an expert code reviewer.

## Your Role
- Review code for bugs and best practices
- Suggest improvements
- Be constructive and helpful
EOF

ayo agents create @reviewer -n -m gpt-4.1 -f system.md
```

### Update Built-in Agents

```bash
# Check for modifications first
ayo agents update

# Force update, overwriting modifications
ayo agents update --force
```

## Skill Management

### List Skills

```bash
ayo skills list
```

Shows all available skills grouped by source.

### Show Skill Details

```bash
ayo skills show skill-name
```

### Create Skill

```bash
# Create in current directory
ayo skills create my-skill

# Create in shared skills directory (~/.config/ayo/skills/)
ayo skills create my-skill --shared

# Create in local project skills (./.config/ayo/skills/)
ayo skills create my-skill --dev
```

This creates a skill directory with a template `SKILL.md` file.

### Validate Skill

```bash
ayo skills validate ./path/to/skill
```

Checks that a skill directory has valid structure and metadata.

### Update Built-in Skills

```bash
ayo skills update
ayo skills update --force
```

## Session Management

Sessions persist conversation history, allowing you to continue previous conversations.

### List Sessions

```bash
# List recent sessions
ayo sessions list

# Filter by agent
ayo sessions list --agent @ayo

# Filter by source (ayo, crush, crush-via-ayo)
ayo sessions list --source crush-via-ayo

# Limit results
ayo sessions list --limit 50
```

### Show Session Details

```bash
# Show session info and messages
ayo sessions show abc123
```

Accepts full session ID or prefix.

### Continue a Session

```bash
# Interactive picker for recent sessions
ayo sessions continue

# Continue specific session by ID prefix
ayo sessions continue abc123

# Search by title
ayo sessions continue "debugging issue"
```

Alias: `ayo sessions resume`

### Delete a Session

```bash
# With confirmation prompt
ayo sessions delete abc123

# Force delete without confirmation
ayo sessions delete abc123 --force
```

## Agent Chaining

Agents with input/output schemas can be chained via Unix pipes.

### List Chainable Agents

```bash
ayo chain ls
ayo chain ls --json
```

### Inspect Agent Schemas

```bash
ayo chain inspect @agent-name
ayo chain inspect @agent-name --json
```

### Find Compatible Agents

```bash
# Agents that can receive this agent's output
ayo chain from @source-agent

# Agents whose output this agent can receive
ayo chain to @target-agent
```

### Validate Input

```bash
# Validate JSON against agent's input schema
ayo chain validate @agent-name '{"key": "value"}'

# Or via stdin
echo '{"key": "value"}' | ayo chain validate @agent-name
```

### Generate Example Input

```bash
ayo chain example @agent-name
```

### Chaining Example

```bash
# Pipe output from one agent to another
ayo @code-reviewer '{"files": ["main.go"]}' | ayo @issue-reporter
```

## Setup and Configuration

### Initial Setup

```bash
# Standard setup
ayo setup

# Force reinstall, overwriting modifications
ayo setup --force

# Development mode (use local directories)
ayo setup --dev
```

### Directory Structure

**Production:**
- User config: `~/.config/ayo/`
- Built-in data: `~/.local/share/ayo/`

**Development (--dev):**
- User config: `./.config/ayo/`
- Built-in data: `./.local/share/ayo/`

### Configuration File

Located at `~/.config/ayo/ayo.json`:

```json
{
  "$schema": "./ayo-schema.json",
  "default_model": "gpt-4.1",
  "provider": {
    "openai": {
      "api_key": "sk-...",
      "models": ["gpt-4.1", "gpt-4.1-mini"]
    }
  }
}
```

## Agent Directory Structure

```
@agent-name/
├── config.json      # Agent configuration
├── system.md        # System prompt
├── input.jsonschema # Optional: input validation schema
├── output.jsonschema # Optional: output format schema
└── skills/          # Optional: agent-specific skills
    └── my-skill/
        └── SKILL.md
```

### config.json Fields

```json
{
  "model": "gpt-4.1",
  "description": "Agent description",
  "allowed_tools": ["bash", "agent_call"],
  "skills": ["skill-a", "skill-b"],
  "exclude_skills": ["unwanted-skill"],
  "ignore_builtin_skills": false,
  "ignore_shared_skills": false,
  "no_system_wrapper": false
}
```

## Skill Directory Structure

```
skill-name/
├── SKILL.md         # Required: skill definition with YAML frontmatter
├── scripts/         # Optional: executable scripts
├── references/      # Optional: additional documentation
└── assets/          # Optional: templates, data files
```

### SKILL.md Format

```markdown
---
name: skill-name
description: What this skill does and when to use it.
license: MIT
metadata:
  author: your-name
  version: "1.0"
---

# Skill Title

Instructions for the agent...
```

## Available Tools

| Tool | Description |
|------|-------------|
| `bash` | Execute shell commands |
| `agent_call` | Delegate tasks to other agents |

## Common Workflows

### Create a Specialized Agent

```bash
# 1. Create system prompt
cat > ~/prompts/debugger.md << 'EOF'
You are an expert debugger. When given an error:
1. Analyze the error message
2. Search for relevant code
3. Identify the root cause
4. Suggest a fix with code examples
EOF

# 2. Create the agent
ayo agents create @debugger -n \
  -m gpt-4.1 \
  -d "Debugging assistant" \
  -f ~/prompts/debugger.md \
  -t bash \
  --skills debugging

# 3. Use it
ayo @debugger "Why is this test failing?"
```

### Create a Chainable Agent

```bash
# Create input schema
cat > input.json << 'EOF'
{
  "type": "object",
  "properties": {
    "code": {"type": "string"},
    "language": {"type": "string"}
  },
  "required": ["code"]
}
EOF

# Create agent with schema
ayo agents create @analyzer -n \
  -m gpt-4.1 \
  -s "Analyze the provided code and return JSON with findings." \
  --input-schema input.json
```

## Error Handling

If a command fails:
1. Check `--help` for correct usage
2. Verify the agent/skill exists with `list` command
3. Check configuration file syntax
4. Run with `--debug` for verbose output
