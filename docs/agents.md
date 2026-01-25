# Agents

Agents are AI assistants with custom system prompts and tool access. Each agent is a directory containing configuration and instructions.

## Built-in Agents

| Agent | Description |
|-------|-------------|
| `@ayo` | Default versatile assistant with bash and tool access |
| `@ayo.agents` | Agent management (creating/modifying agents) |
| `@ayo.skills` | Skill management (creating/modifying skills) |

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
# Default agent
ayo

# Specific agent
ayo @ayo

# With file attachment
ayo @ayo -a main.go
```

Exit with `Ctrl+C` (twice if mid-response).

### Single Prompt

```bash
# Run prompt and exit
ayo "what's new in Go 1.22?"

# With specific agent
ayo @ayo "explain this error"

# With file attachments
ayo -a error.log "what caused this?"
```

## Creating Agents

### Interactive Wizard

```bash
ayo agents create @myagent
```

The wizard guides you through:
1. Model selection
2. Description
3. System prompt (via editor)
4. Tools selection
5. Skills selection

### Non-Interactive

```bash
ayo agents create @myagent \
  --non-interactive \
  --model gpt-4.1 \
  --description "My custom agent" \
  --system "You are a helpful assistant..." \
  --tools bash,agent_call \
  --skills debugging
```

### Using a System Prompt File

```bash
# Create system prompt
cat > system.md << 'EOF'
You are an expert code reviewer.

## Guidelines
- Review for bugs and best practices
- Suggest improvements
- Be constructive
EOF

# Create agent with file
ayo agents create @reviewer -n -m gpt-4.1 -f system.md
```

### Create Flags

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
| `--non-interactive` | `-n` | Skip wizard |

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

### config.json

```json
{
  "description": "What this agent does",
  "model": "gpt-4.1",
  "allowed_tools": ["bash", "agent_call"],
  "skills": ["debugging", "coding"],
  "exclude_skills": ["unwanted-skill"],
  "ignore_builtin_skills": false,
  "ignore_shared_skills": false,
  "guardrails": true,
  "delegates": {
    "coding": "@crush"
  }
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Brief agent description |
| `model` | string | LLM model (uses default if omitted) |
| `allowed_tools` | string[] | Tools the agent can use |
| `skills` | string[] | Skills to attach |
| `exclude_skills` | string[] | Skills to exclude |
| `ignore_builtin_skills` | bool | Skip built-in skills |
| `ignore_shared_skills` | bool | Skip user shared skills |
| `guardrails` | bool | Safety guardrails (default: true) |
| `delegates` | object | Task type to agent mappings |

### system.md

The system prompt defines the agent's behavior:

```markdown
You are an expert debugger.

## Your Role

- Analyze error messages
- Search for relevant code
- Identify root causes
- Suggest fixes with code examples

## Guidelines

1. Be systematic and thorough
2. Explain your reasoning
3. Provide actionable solutions
```

## Managing Agents

### List Agents

```bash
ayo agents list
```

Shows agents grouped by source (user-defined vs built-in).

### Show Details

```bash
ayo agents show @ayo
```

Displays configuration, tools, skills, and location.

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
ayo agents create @dangerous -n --no-guardrails
```

## System Prompt Assembly

When an agent runs, the system prompt is assembled from multiple sources:

1. **Environment context** - Platform, date, git status
2. **Guardrails** - If enabled
3. **User prefix** - Optional `~/.config/ayo/prompts/system-prefix.md`
4. **Agent prompt** - From `system.md`
5. **User suffix** - Optional `~/.config/ayo/prompts/system-suffix.md`
6. **Tools prompt** - Tool instructions
7. **Skills prompt** - Available skills

## Reserved Namespaces

The `@ayo` namespace is reserved for built-in agents:
- Users cannot create `@ayo` or `@ayo.*` agents
- All `@ayo` agents have guardrails enforced
- Attempting to create `@ayo.custom` will fail

## Delegation

Agents can delegate specific task types to other agents. See [Delegation](delegation.md) for details.

```json
{
  "delegates": {
    "coding": "@crush"
  }
}
```

## Agent Chaining

Agents with I/O schemas can be composed via Unix pipes. See [Chaining](chaining.md) for details.

```bash
ayo @analyzer '{"code":"..."}' | ayo @reporter
```
