# Agents

Agents are AI assistants with custom system prompts and tool access. Each agent is a directory containing configuration and instructions.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Agent                                │
├─────────────────────────────────────────────────────────────┤
│  config.json       Model, tools, skills configuration       │
│  system.md         System prompt (behavior instructions)    │
│  skills/           Agent-specific skills (optional)         │
│  *.jsonschema      Structured I/O for chaining (optional)   │
├─────────────────────────────────────────────────────────────┤
│  Tools             bash, agent_call, plan                   │
│  Skills            Instruction sets loaded at runtime       │
│  Guardrails        Safety constraints (enabled by default)  │
└─────────────────────────────────────────────────────────────┘
```

## Built-in Agent

| Agent | Description |
|-------|-------------|
| `@ayo` | Default versatile assistant with bash and tool access |

`@ayo` is designed to handle all tasks including agent and skill management. To create or manage agents, just ask:

```bash
ayo "help me create an agent for code review"
ayo "show me what agents I have"
```

### Agents via Plugins

Additional agents are available through plugins:

| Agent | Plugin | Description |
|-------|--------|-------------|
| `@research` | ayo-plugins-research | Research assistant with web search |
| `@crush` | ayo-plugins-crush | Coding agent powered by Crush |

```bash
# Install a plugin to get additional agents
ayo plugins install https://github.com/user/ayo-plugins-research
```

## Running Agents

### Interactive Chat

```bash
# Default agent (@ayo)
ayo

# Specific agent
ayo @myagent

# With file attachment
ayo -a main.go
```

Exit with `Ctrl+C` (twice if mid-response).

### Single Prompt

```bash
# Run prompt and exit
ayo "what's new in Go 1.22?"

# With specific agent
ayo @myagent "explain this error"

# With file attachments
ayo -a error.log "what caused this?"
```

## Creating Agents

### Conversational Approach (Recommended)

Ask `@ayo` to help design and create your agent:

```bash
ayo "help me create an agent for code review"
```

This provides a guided experience where `@ayo` will:
1. Ask about the agent's purpose
2. Help design the system prompt
3. Suggest appropriate tools and skills
4. Create the agent files

### CLI Approach

For scripted or quick creation:

```bash
ayo agents create @reviewer \
  -m gpt-5.2 \
  -d "Reviews code for best practices" \
  -f ~/prompts/reviewer.md \
  -t bash \
  --skills debugging
```

### Using a System Prompt File

```bash
# Create system prompt
cat > system.md << 'EOF'
# Code Reviewer

You are an expert code reviewer.

## Guidelines
- Review for bugs and best practices
- Suggest improvements
- Be constructive
EOF

# Create agent with file
ayo agents create @reviewer -m gpt-5.2 -f system.md
```

### Create Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--model` | `-m` | Model to use |
| `--description` | `-d` | Brief description |
| `--system` | `-s` | System prompt text (inline) |
| `--system-file` | `-f` | Path to system prompt file |
| `--tools` | `-t` | Allowed tools (comma-separated) |
| `--skills` | | Skills to include |
| `--exclude-skills` | | Skills to exclude |
| `--ignore-builtin-skills` | | Don't load built-in skills |
| `--ignore-shared-skills` | | Don't load user shared skills |
| `--input-schema` | | JSON schema for stdin input |
| `--output-schema` | | JSON schema for stdout output |
| `--no-guardrails` | | Disable safety guardrails |

## Agent Structure

```
@myagent/
├── config.json         # Agent configuration
├── system.md           # System prompt
├── skills/             # Agent-specific skills (optional)
│   └── my-skill/
│       └── SKILL.md
├── input.jsonschema    # Input schema for chaining (optional)
└── output.jsonschema   # Output schema for chaining (optional)
```

### Locations

| Location | Path | Purpose |
|----------|------|---------|
| User agents | `~/.config/ayo/agents/` | Your custom agents |
| Built-in | `~/.local/share/ayo/agents/` | Shipped with ayo |

User agents take precedence over built-in agents with the same name.

### config.json

```json
{
  "description": "What this agent does",
  "model": "gpt-5.2",
  "allowed_tools": ["bash", "agent_call"],
  "skills": ["debugging"],
  "exclude_skills": [],
  "ignore_builtin_skills": false,
  "ignore_shared_skills": false,
  "guardrails": true,
  "delegates": {
    "coding": "@crush"
  }
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | | Brief agent description |
| `model` | string | (global) | LLM model to use |
| `allowed_tools` | string[] | `["bash"]` | Tools the agent can use |
| `skills` | string[] | `[]` | Skills to attach |
| `exclude_skills` | string[] | `[]` | Skills to exclude |
| `ignore_builtin_skills` | bool | `false` | Skip built-in skills |
| `ignore_shared_skills` | bool | `false` | Skip user shared skills |
| `guardrails` | bool | `true` | Safety guardrails |
| `delegates` | object | | Task type to agent mappings |

### system.md

The system prompt defines the agent's behavior:

```markdown
# Expert Debugger

You are an expert debugger specializing in finding and fixing bugs.

## Your Role

- Analyze error messages
- Search for relevant code
- Identify root causes
- Suggest fixes with code examples

## Guidelines

1. Be systematic and thorough
2. Explain your reasoning
3. Provide actionable solutions

## Output Format

For each bug found:
1. Location (file and line)
2. Issue description
3. Root cause analysis
4. Suggested fix
```

## Managing Agents

### List Agents

```bash
ayo agents list
```

Shows agents grouped by source (user-defined vs built-in).

### Show Details

```bash
ayo agents show @myagent
```

Displays configuration, tools, skills, and location.

### Edit an Agent

```bash
# Edit system prompt
$EDITOR ~/.config/ayo/agents/@myagent/system.md

# Edit configuration
$EDITOR ~/.config/ayo/agents/@myagent/config.json
```

### Delete an Agent

```bash
rm -rf ~/.config/ayo/agents/@myagent
```

### Update Built-ins

```bash
# Check for modifications first
ayo agents update

# Force update (overwrites modifications)
ayo agents update --force
```

## Guardrails

Guardrails are safety constraints applied to agent system prompts. They enforce:
- No malicious code creation
- No credential exposure
- Confirmation before destructive actions
- Scope limitation to current project

### Configuration

Guardrails are **enabled by default**. To disable (not recommended):

```json
{
  "guardrails": false
}
```

**Note:** Agents in the `@ayo` namespace always have guardrails enabled regardless of this setting.

### CLI Flag

```bash
# Create agent without guardrails (dangerous)
ayo agents create @dangerous --no-guardrails
```

## System Prompt Assembly

When an agent runs, the system prompt is assembled from multiple sources:

```
┌─────────────────────────────────────┐
│  1. Environment context             │  Platform, date, git status
│  2. Guardrails                      │  Safety constraints (if enabled)
│  3. User prefix                     │  ~/.config/ayo/prompts/prefix.md
│  4. Agent system prompt             │  system.md
│  5. User suffix                     │  ~/.config/ayo/prompts/suffix.md
│  6. Tools prompt                    │  Tool instructions
│  7. Skills prompt                   │  Attached skill instructions
└─────────────────────────────────────┘
```

## Reserved Namespaces

The `@ayo` namespace is reserved for built-in agents:
- Users cannot create `@ayo` or `@ayo.*` agents
- All `@ayo` agents have guardrails enforced
- Attempting to create `@ayo.custom` will fail

## Available Tools

| Tool | Description |
|------|-------------|
| `bash` | Execute shell commands |
| `agent_call` | Delegate tasks to other agents |
| `memory` | Store and search persistent memories |
| `todo` | Track multi-step tasks with a todo list |
| `search` | Web search (requires configured provider) |

## Delegation

Agents can delegate specific task types to other agents. See [Delegation](delegation.md) for details.

```json
{
  "delegates": {
    "coding": "@crush",
    "research": "@research"
  }
}
```

## Agent Chaining

Agents with I/O schemas can be composed via Unix pipes. See [Chaining](chaining.md) for details.

```bash
ayo @analyzer '{"code":"..."}' | ayo @reporter
```
