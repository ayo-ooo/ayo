---
name: ayo
description: Manage ayo agents, skills, and configuration using the ayo CLI. Use when the user wants to create, list, modify, or understand agents and skills.
metadata:
  author: ayo
  version: "2.0"
---

# Ayo CLI Skill

This skill provides comprehensive instructions for using the ayo command-line interface to manage agents, skills, and configuration.

## CRITICAL: Always Use the CLI

When managing agents or skills, you MUST use the `ayo` CLI commands via bash. NEVER write directly to ayo's directories.

**CORRECT - Use CLI commands:**
```bash
ayo agents create @myagent -m gpt-5.2 -f /tmp/system.md
ayo agents list
ayo skills create myskill --shared
```

**WRONG - Never do this:**
```bash
mkdir -p ~/.config/ayo/agents/@myagent  # WRONG
cat > ~/.config/ayo/agents/@myagent/config.json  # WRONG
mkdir custom_agents  # WRONG
```

The CLI handles proper directory structure, validation, and installation.

## When to Use

Activate this skill when:
- User wants to create, edit, or delete agents
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
| `ayo flows` | Manage flows (list, run, history, replay) |
| `ayo plugins` | Manage plugins (install, list, update, remove) |
| `ayo sessions` | Manage conversation sessions |
| `ayo memory` | Manage agent memories |
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

---

# Agent Management

## List Agents

```bash
ayo agents list
```

Shows all available agents grouped by source (user-defined vs built-in).

## Show Agent Details

```bash
ayo agents show @agent-name
```

Displays agent configuration including model, tools, skills, and location.

## Create Agent

Non-interactive (recommended for scripted creation):
```bash
ayo agents create @my-agent \
  --non-interactive \
  --model gpt-5.2 \
  --description "My agent description" \
  --system-file system.md \
  --tools bash \
  --skills debugging
```

### Create Agent Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--model` | `-m` | Model to use (required in non-interactive mode) |
| `--description` | `-d` | Brief description of the agent |
| `--system` | `-s` | System prompt text (inline) |
| `--system-file` | `-f` | Path to system prompt file (.md or .txt) |
| `--tools` | `-t` | Allowed tools: bash, agent_call, todo (comma-separated) |
| `--skills` | | Skills to include (comma-separated) |
| `--exclude-skills` | | Skills to exclude |
| `--ignore-builtin-skills` | | Don't load built-in skills |
| `--ignore-shared-skills` | | Don't load user shared skills |
| `--input-schema` | | Path to JSON schema for validating stdin input |
| `--output-schema` | | Path to JSON schema for structuring stdout output |
| `--no-guardrails` | | Disable system guardrails (not recommended) |

## Update Built-in Agents

```bash
# Check for modifications first
ayo agents update

# Force update, overwriting modifications
ayo agents update --force
```

## Edit an Agent

User agents are stored in `~/.config/ayo/agents/@{name}/`. Edit files directly:

```bash
# Edit system prompt
$EDITOR ~/.config/ayo/agents/@my-agent/system.md

# Edit configuration
$EDITOR ~/.config/ayo/agents/@my-agent/config.json
```

**Programmatic edits:** Use bash to modify agent files:

```bash
# Append to system prompt
cat >> ~/.config/ayo/agents/@my-agent/system.md << 'EOF'

## Additional Instructions
Write output to a markdown file in the current working directory.
EOF

# Update config.json (add a tool)
# Read current config, modify, and write back
```

## Update vs Create Decision

**CRITICAL:** When a user refers to an existing agent and requests changes, **update the existing agent** rather than creating a new one.

### Decision Flow

1. **Check if agent exists**: `ayo agents show @agent-name`
2. **If exists + user wants changes** → Edit the existing agent's files
3. **If doesn't exist** → Create new agent

### Signals to Update (not create):
- User says "I want **it** to..." (refers to previously discussed agent)
- User says "change the agent to..." or "update it to..."
- User says "add X to the agent" or "make it also do Y"
- Context shows an agent was just created or discussed

### Signals to Create:
- User explicitly says "create a new agent" or "make me an agent"
- No prior agent context in conversation
- User wants a completely different agent type

### Example: Update Flow

User just created `@researcher`, then says: "I want it to write output to markdown files"

**Correct approach:**
```bash
# Append to the existing system prompt
cat >> ~/.config/ayo/agents/@researcher/system.md << 'EOF'

## Output Format
Always write research findings to a markdown file named `<topic>.md` in the current working directory.
EOF
```

**Wrong approach:** Creating a new `@researcher2` or `@markdown-researcher` agent.

## Remove an Agent

```bash
rm -rf ~/.config/ayo/agents/@agent-name
```

---

# Creating Effective Agents

## Agent Directory Structure

```
@agent-name/
├── config.json         # Required: Agent configuration
├── system.md           # Required: System prompt
├── input.jsonschema    # Optional: Input validation schema (for chaining)
├── output.jsonschema   # Optional: Output format schema (for chaining)
└── skills/             # Optional: Agent-specific skills
    └── my-skill/
        └── SKILL.md
```

### Agent Installation Locations

| Location | Path | When to Use |
|----------|------|-------------|
| User agents | `~/.config/ayo/agents/` | Personal agents available everywhere |
| Dev mode | `./.config/ayo/agents/` | Project-specific or testing agents |

## config.json Reference

```json
{
  "model": "gpt-5.2",
  "description": "What this agent does",
  "allowed_tools": ["bash", "agent_call", "todo"],
  "skills": ["skill-a", "skill-b"],
  "exclude_skills": ["unwanted-skill"],
  "ignore_builtin_skills": false,
  "ignore_shared_skills": false,
  "guardrails": true
}
```

### Field Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | (global default) | LLM model to use |
| `description` | string | | Brief description shown in `ayo agents list` |
| `allowed_tools` | array | `["bash"]` | Tools: `bash`, `agent_call`, `plan` |
| `skills` | array | `[]` | Skills to load for this agent |
| `exclude_skills` | array | `[]` | Skills to explicitly exclude |
| `ignore_builtin_skills` | bool | `false` | Don't load any built-in skills |
| `ignore_shared_skills` | bool | `false` | Don't load user shared skills |
| `guardrails` | bool | `true` | Safety guardrails (set false to disable - dangerous) |

### Configuration Patterns

**Minimal agent** (just bash):
```json
{
  "description": "Simple automation agent",
  "allowed_tools": ["bash"]
}
```

**Agent with delegation** (can call other agents):
```json
{
  "description": "Orchestrator agent",
  "allowed_tools": ["bash", "agent_call"]
}
```

**Agent with todo tracking** (tracks multi-step tasks):
```json
{
  "description": "Project planner",
  "allowed_tools": ["bash", "todo"]
}
```

**Skill-focused agent** (specific expertise):
```json
{
  "description": "Debugging specialist",
  "allowed_tools": ["bash"],
  "skills": ["debugging"],
  "ignore_builtin_skills": true
}
```

---

# Tool Discovery and Selection

**CRITICAL:** Before creating an agent, always discover what tools are available and proactively select the appropriate ones based on the agent's intended purpose.

## Discovering Available Tools

### Built-in Tools

| Tool | Purpose | When to Include |
|------|---------|-----------------|
| `bash` | Execute shell commands | Almost always (default) |
| `plan` | Track multi-step tasks with phases/todos | Complex workflows, project management |
| `agent_call` | Delegate to other agents | Orchestrators, managers, routers |
| `memory` | Store/retrieve persistent facts | Personalization, learning agents |
| `search` | Web search (if configured) | Research, information gathering |

### Discovering Plugin Tools

```bash
# List installed plugins and their tools
ayo plugins list

# Show tools from a specific plugin
ayo plugins show <plugin-name>
```

### Checking Default Tool Mappings

Plugins can provide default tool mappings (e.g., `search` → `searxng`). Check what's configured:

```bash
cat ~/.config/ayo/ayo.json | grep -A5 default_tools
```

If a default is set (e.g., `"search": "searxng"`), agents can use `search` in their `allowed_tools` and it resolves to the plugin tool automatically.

## Proactive Tool Selection

When a user describes an agent's purpose, **analyze the requirements and suggest appropriate tools**:

| Agent Purpose | Recommended Tools | Reasoning |
|---------------|-------------------|-----------|
| Research / information gathering | `bash`, `search` | Needs web search capability |
| Code development | `bash`, `plan` | Needs execution + task tracking |
| Orchestration / delegation | `bash`, `agent_call` | Needs to call other agents |
| Long-running projects | `bash`, `plan`, `memory` | Task tracking + persistence |
| Simple automation | `bash` | Just command execution |
| Personalized assistant | `bash`, `memory` | Needs to remember preferences |

### Decision Flow

When creating an agent:

1. **Identify capabilities needed** from the agent's description
2. **Check available tools** with `ayo plugins list` and `ayo plugins show`
3. **Check default mappings** in global config
4. **Include all required tools** in `allowed_tools`
5. **Inform the user** if a needed tool isn't installed

### Example: Research Agent

User wants: "Create a deep research agent that searches the web"

**Before creating, check:**
```bash
# Is search available?
ayo plugins list
cat ~/.config/ayo/ayo.json | grep -A5 default_tools
```

**If search is configured:**
```bash
ayo agents create @deep-research -n \
  -m gpt-5.2 \
  -d "Deep research agent with web search" \
  -t bash,search \
  -f system.md
```

**If search is NOT available:**
Inform the user: "A search tool is recommended for research agents. Install a search plugin (e.g., `ayo plugins install <search-plugin-url>`) to enable web search capabilities."

---

# System Prompt Design

A well-crafted system prompt is the most important part of an agent. Follow this template:

## System Prompt Template

Create a `system.md` file with this structure:

```markdown
# {Role Title}

You are {specific role description}. {Additional context about the role and expertise}.

## Your Responsibilities

- {Primary task or capability}
- {Secondary task or capability}
- {Additional tasks as needed}

## Guidelines

- {Important behavioral guideline}
- {Quality standard or constraint}
- {Communication style preference}

## Workflow

When given a task:
1. {First step in your process}
2. {Second step}
3. {Continue as needed}

## Output Format

{Describe how responses should be structured - format, length, style}

## Examples

### Example 1: {Scenario name}

**User request:** {example input}

**Your response:** {example of ideal response}

### Example 2: {Another scenario}

**User request:** {example input}

**Your response:** {example of ideal response}
```

## System Prompt Best Practices

1. **Define identity clearly**: Start with "You are..." to establish role and expertise
2. **Set explicit boundaries**: What the agent should and shouldn't do
3. **Be specific about tasks**: Vague instructions lead to vague results
4. **Include examples**: Show input/output pairs for complex behaviors
5. **Define output format**: How should responses be structured?
6. **Keep it focused**: One agent = one clear purpose

## Example System Prompts

### Code Reviewer Agent

```markdown
# Code Reviewer

You are an expert code reviewer specializing in identifying bugs, security issues, and opportunities for improvement.

## Your Responsibilities

- Review code for bugs, logic errors, and edge cases
- Identify security vulnerabilities
- Suggest performance improvements
- Ensure code follows best practices
- Be constructive and educational

## Guidelines

- Always explain WHY something is an issue, not just what
- Prioritize issues by severity (critical > high > medium > low)
- Suggest specific fixes, not just problems
- Acknowledge good patterns when you see them

## Output Format

For each issue found:
1. **Location**: File and line number
2. **Severity**: Critical/High/Medium/Low
3. **Issue**: What's wrong
4. **Why**: Why it matters
5. **Fix**: Suggested solution

## Example

**Input:** Review `auth.go` lines 45-60

**Output:**
### Critical: SQL Injection Vulnerability
**Location:** auth.go:52
**Issue:** User input concatenated directly into SQL query
**Why:** Allows attackers to execute arbitrary SQL commands
**Fix:** Use parameterized queries:
\`\`\`go
db.Query("SELECT * FROM users WHERE id = ?", userID)
\`\`\`
```

### Research Agent

```markdown
# Research Assistant

You are a thorough research assistant who gathers, synthesizes, and presents information clearly.

## Your Responsibilities

- Research topics using available tools
- Synthesize information from multiple sources
- Present findings in a structured format
- Cite sources when possible
- Distinguish between facts and speculation

## Guidelines

- Be thorough but focused on the user's question
- Present multiple perspectives when relevant
- Acknowledge uncertainty or gaps in information
- Keep summaries concise but complete

## Output Format

1. **Summary**: 2-3 sentence overview
2. **Key Findings**: Bullet points of main discoveries
3. **Details**: Elaboration on important points
4. **Sources**: Where information came from (if applicable)
```

### Task Automation Agent

```markdown
# Task Automator

You are an automation specialist who executes system tasks efficiently and safely.

## Your Responsibilities

- Execute file system operations
- Run build and test commands
- Manage dependencies
- Automate repetitive tasks

## Guidelines

- Always verify before destructive operations
- Use dry-run flags when available
- Report what was done, not what you're about to do
- Handle errors gracefully

## Workflow

1. Understand the task requirements
2. Plan the sequence of commands
3. Execute commands one at a time
4. Verify success before proceeding
5. Report results concisely

## Output Format

After completing a task:
- What was done (past tense)
- Any notable outcomes or warnings
- Next steps if applicable
```

---

# Agent Chaining with Structured I/O

Agents with input/output schemas can be chained via Unix pipes. This enables powerful multi-step workflows.

## When to Use Schemas

| Use Case | Input Schema | Output Schema |
|----------|--------------|---------------|
| Agent receives structured data | Yes | Optional |
| Agent produces structured data | Optional | Yes |
| Agent is part of a pipeline | Yes | Yes |
| General conversational agent | No | No |

## Schema Basics

Schemas use JSON Schema format. Key concepts:

| Keyword | Purpose | Example |
|---------|---------|---------|
| `type` | Data type | `"string"`, `"object"`, `"array"`, `"integer"`, `"boolean"` |
| `properties` | Object fields | `{"name": {"type": "string"}}` |
| `required` | Mandatory fields | `["name", "email"]` |
| `items` | Array element schema | `{"type": "string"}` |
| `description` | Human-readable docs | `"User's email address"` |
| `enum` | Allowed values | `["low", "medium", "high"]` |

## Input Schema (input.jsonschema)

Defines what the agent accepts as input. When an input schema exists:
- Agent only accepts JSON matching this schema
- Input is validated before processing
- User sees helpful error if input doesn't match

### Input Schema Template

```json
{
  "type": "object",
  "properties": {
    "required_field": {
      "type": "string",
      "description": "Description of this field"
    },
    "optional_field": {
      "type": "integer",
      "description": "Optional configuration"
    },
    "array_field": {
      "type": "array",
      "items": { "type": "string" },
      "description": "List of items"
    },
    "nested_object": {
      "type": "object",
      "properties": {
        "sub_field": { "type": "boolean" }
      }
    }
  },
  "required": ["required_field"]
}
```

### Input Schema Examples

**File analyzer input:**
```json
{
  "type": "object",
  "properties": {
    "files": {
      "type": "array",
      "items": { "type": "string" },
      "description": "List of file paths to analyze"
    },
    "options": {
      "type": "object",
      "properties": {
        "verbose": {
          "type": "boolean",
          "description": "Include detailed output"
        },
        "format": {
          "type": "string",
          "enum": ["json", "markdown", "text"],
          "description": "Output format"
        }
      }
    }
  },
  "required": ["files"]
}
```

**Code review input:**
```json
{
  "type": "object",
  "properties": {
    "code": {
      "type": "string",
      "description": "Source code to review"
    },
    "language": {
      "type": "string",
      "description": "Programming language"
    },
    "context": {
      "type": "string",
      "description": "Additional context about the code"
    }
  },
  "required": ["code"]
}
```

## Output Schema (output.jsonschema)

Defines the structure of the agent's output. When an output schema exists:
- Agent's final response is formatted as JSON matching this schema
- Enables piping to downstream agents
- Provides consistent, parseable output

### Output Schema Template

```json
{
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["success", "error", "partial"],
      "description": "Overall result status"
    },
    "data": {
      "type": "object",
      "description": "Main output data"
    },
    "summary": {
      "type": "string",
      "description": "Human-readable summary"
    },
    "errors": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "code": { "type": "string" },
          "message": { "type": "string" }
        }
      }
    }
  },
  "required": ["status"]
}
```

### Output Schema Examples

**Code review findings:**
```json
{
  "type": "object",
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "file": { "type": "string" },
          "line": { "type": "integer" },
          "severity": {
            "type": "string",
            "enum": ["critical", "high", "medium", "low"]
          },
          "category": {
            "type": "string",
            "enum": ["bug", "security", "performance", "style"]
          },
          "message": { "type": "string" },
          "suggestion": { "type": "string" }
        },
        "required": ["file", "severity", "message"]
      }
    },
    "summary": {
      "type": "object",
      "properties": {
        "total_issues": { "type": "integer" },
        "by_severity": {
          "type": "object",
          "properties": {
            "critical": { "type": "integer" },
            "high": { "type": "integer" },
            "medium": { "type": "integer" },
            "low": { "type": "integer" }
          }
        }
      }
    }
  }
}
```

**Analysis result:**
```json
{
  "type": "object",
  "properties": {
    "analysis": {
      "type": "object",
      "properties": {
        "complexity": {
          "type": "string",
          "enum": ["low", "medium", "high"]
        },
        "maintainability_score": {
          "type": "number",
          "minimum": 0,
          "maximum": 100
        },
        "recommendations": {
          "type": "array",
          "items": { "type": "string" }
        }
      }
    },
    "metrics": {
      "type": "object",
      "properties": {
        "lines_of_code": { "type": "integer" },
        "functions": { "type": "integer" },
        "dependencies": { "type": "integer" }
      }
    }
  }
}
```

## Schema Compatibility

When piping agents, schemas are checked:

| Compatibility | Description |
|---------------|-------------|
| **Exact** | Output schema identical to input schema |
| **Structural** | Output has all required input fields (superset OK) |
| **Freeform** | Target has no input schema (accepts anything) |

## Chain Commands

```bash
# List chainable agents (have schemas)
ayo chain ls

# Inspect agent schemas
ayo chain inspect @agent-name

# Find agents that can receive output
ayo chain from @source-agent

# Find agents that can feed into this one
ayo chain to @target-agent

# Validate input against schema
ayo chain validate @agent-name '{"key": "value"}'

# Generate example input
ayo chain example @agent-name
```

## Creating a Chainable Agent Workflow

```bash
# 1. Create input schema
cat > input.jsonschema << 'EOF'
{
  "type": "object",
  "properties": {
    "code": { "type": "string", "description": "Code to analyze" },
    "language": { "type": "string", "description": "Programming language" }
  },
  "required": ["code"]
}
EOF

# 2. Create output schema
cat > output.jsonschema << 'EOF'
{
  "type": "object",
  "properties": {
    "issues": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "line": { "type": "integer" },
          "message": { "type": "string" }
        }
      }
    }
  }
}
EOF

# 3. Create system prompt
cat > system.md << 'EOF'
# Code Analyzer

You analyze code and return issues as structured JSON matching your output schema.

When given code:
1. Analyze for bugs, issues, and improvements
2. Return findings in the specified JSON format
3. Include line numbers when possible
EOF

# 4. Create the agent
ayo agents create @analyzer -n \
  -m gpt-5.2 \
  -d "Analyzes code and returns structured findings" \
  -f system.md \
  --input-schema input.jsonschema \
  --output-schema output.jsonschema

# 5. Test
ayo chain validate @analyzer '{"code": "print(x)", "language": "python"}'
ayo @analyzer '{"code": "print(x)", "language": "python"}'
```

---

# Skill Management

## List Skills

```bash
ayo skills list
```

Filter by source:
```bash
ayo skills list --source=built-in
ayo skills list --source=shared
ayo skills list --source=local
```

## Show Skill Details

```bash
ayo skills show skill-name
```

## Create Skill

```bash
# Create in current directory
ayo skills create my-skill

# Create in shared skills directory (~/.config/ayo/skills/)
ayo skills create my-skill --shared

# Create in local project skills (./.config/ayo/skills/)
ayo skills create my-skill --dev
```

## Validate Skill

```bash
ayo skills validate ./path/to/skill
```

## Skill Directory Structure

```
skill-name/
├── SKILL.md           # Required: skill definition with YAML frontmatter
├── scripts/           # Optional: executable scripts
├── references/        # Optional: additional documentation
└── assets/            # Optional: templates, data files
```

## SKILL.md Format

```markdown
---
name: skill-name
description: >
  What this skill does and when to use it.
  This appears in skill listings and helps agents
  understand when to apply this skill.
metadata:
  author: your-name
  version: "1.0"
---

# Skill Title

Instructions for the agent when this skill is active...
```

### Required Fields

| Field | Requirements |
|-------|-------------|
| `name` | 1-64 chars, lowercase, hyphens allowed, must match directory name |
| `description` | 1-1024 chars, explains what the skill does and when to use it |

### Optional Fields

| Field | Purpose |
|-------|---------|
| `compatibility` | Environment requirements (max 500 chars) |
| `metadata` | Key-value pairs (author, version, etc.) |
| `allowed-tools` | Pre-approved tools (experimental) |

## Skill Discovery Priority

1. **Agent-specific** (in agent's `skills/` directory)
2. **Project-local** (current working directory)
3. **User shared** (`~/.config/ayo/skills/`)
4. **Built-in** (`~/.local/share/ayo/skills/`)

---

# Session Management

Sessions persist conversation history.

```bash
# List recent sessions
ayo sessions list

# Filter by agent
ayo sessions list --agent @ayo

# Show session details
ayo sessions show abc123

# Continue a session (interactive picker)
ayo sessions continue

# Continue specific session
ayo sessions continue abc123

# Delete a session
ayo sessions delete abc123
```

---

# Memory Management

Memories are persistent facts and preferences learned across sessions.

```bash
# List memories
ayo memory list
ayo memory list --agent @ayo

# Semantic search
ayo memory search "coding preferences"

# Show memory details
ayo memory show abc123

# Forget a memory
ayo memory forget abc123

# Show statistics
ayo memory stats

# Clear all memories
ayo memory clear
```

---

# Plugin Management

Plugins extend ayo with additional agents, skills, and tools.

```bash
# Install from GitHub
ayo plugins install owner/name
ayo plugins install https://github.com/owner/ayo-plugins-name

# List installed plugins
ayo plugins list

# Update plugins
ayo plugins update

# Remove a plugin
ayo plugins remove <name>
```

---

# Flow Management

Flows are composable agent pipelines - shell scripts with structured frontmatter that orchestrate agent calls.

## List Flows

```bash
# List all flows
ayo flows list

# Filter by source
ayo flows list --source=project
ayo flows list --source=user
ayo flows list --source=built-in

# JSON output
ayo flows list --json
```

## Show Flow Details

```bash
# Show flow details
ayo flows show my-flow

# Show full script content
ayo flows show my-flow --script

# JSON output
ayo flows show my-flow --json
```

## Run a Flow

```bash
# Run with inline JSON input
ayo flows run my-flow '{"key": "value"}'

# Run with input from stdin
echo '{"key": "value"}' | ayo flows run my-flow

# Run with input from file
ayo flows run my-flow -i input.json

# Custom timeout (seconds)
ayo flows run my-flow -t 600 '{"key": "value"}'

# Validate input only (don't execute)
ayo flows run my-flow --validate '{"key": "value"}'

# Skip history recording
ayo flows run my-flow --no-history '{"key": "value"}'
```

### Run Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--input` | `-i` | Input file path |
| `--timeout` | `-t` | Timeout in seconds (default 300) |
| `--validate` | | Validate input only, don't run |
| `--no-history` | | Don't record run in history |

## Create a Flow

```bash
# Create in user flows directory
ayo flows new my-flow

# Create in project directory (.ayo/flows/)
ayo flows new my-flow --project

# Create with input/output schemas
ayo flows new my-flow --with-schemas

# Overwrite existing
ayo flows new my-flow --force
```

## Validate a Flow

```bash
# Validate flow file
ayo flows validate /path/to/flow.sh

# Validate flow directory (with schemas)
ayo flows validate /path/to/flow-dir/
```

## Flow History

```bash
# List recent runs
ayo flows history

# Filter by flow name
ayo flows history --flow=my-flow

# Filter by status
ayo flows history --status=failed
ayo flows history --status=success
ayo flows history --status=timeout

# Limit results
ayo flows history --limit=20

# JSON output
ayo flows history --json

# Show specific run details
ayo flows history show <run-id>
ayo flows history show <run-id> --json
```

## Replay a Flow

```bash
# Replay a previous run with original input
ayo flows replay <run-id>

# Custom timeout
ayo flows replay <run-id> -t 600

# Skip history recording
ayo flows replay <run-id> --no-history
```

## Flow File Format

Flows are shell scripts with frontmatter:

```bash
#!/usr/bin/env bash
# ayo:flow
# name: my-flow
# description: What this flow does
# version: 1.0.0
# author: username

set -euo pipefail

INPUT="${1:-$(cat)}"

# Process through agents
echo "$INPUT" | ayo @ayo "Process this and return JSON"
```

### Required Frontmatter

| Field | Description |
|-------|-------------|
| `# ayo:flow` | Marker identifying this as a flow |
| `# name:` | Flow name (lowercase, hyphens allowed) |
| `# description:` | What the flow does |

### Optional Frontmatter

| Field | Description |
|-------|-------------|
| `# version:` | Semantic version |
| `# author:` | Author name |

## Flow Directories

| Location | Path | Priority |
|----------|------|----------|
| Project | `.ayo/flows/` | 1 (highest) |
| User | `~/.config/ayo/flows/` | 2 |
| Built-in | `~/.local/share/ayo/flows/` | 3 |

## Structured I/O

Flows can define JSON schemas for type-safe input/output:

```
my-flow/
├── flow.sh              # The flow script
├── input.jsonschema     # Input validation schema
└── output.jsonschema    # Output validation schema
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Input validation failed |
| 124 | Timeout |

## Environment Variables

During execution, flows have access to:

| Variable | Description |
|----------|-------------|
| `AYO_FLOW_NAME` | Name of the flow |
| `AYO_FLOW_RUN_ID` | Unique run ID (ULID) |
| `AYO_FLOW_DIR` | Directory containing the flow |
| `AYO_FLOW_INPUT_FILE` | Path to input file (for large inputs) |

---

# Configuration

## Configuration File

Located at `~/.config/ayo/ayo.json`:

```json
{
  "$schema": "./ayo-schema.json",
  "default_model": "gpt-5.2",
  "provider": {
    "openai": {
      "api_key": "sk-...",
      "models": ["gpt-5.2", "gpt-5.2-mini"]
    }
  }
}
```

## Directory Structure

**Production:**
- User config: `~/.config/ayo/`
- Built-in data: `~/.local/share/ayo/`

**Development (--dev):**
- User config: `./.config/ayo/`
- Built-in data: `./.local/share/ayo/`

---

# Troubleshooting

## Agent Issues

| Problem | Cause | Solution |
|---------|-------|----------|
| Agent gives generic responses | System prompt too vague | Add specific instructions and examples |
| Agent doesn't use tools | Tools not in `allowed_tools` | Add required tools to config.json |
| Agent can't use plugin tool | Tool not in `allowed_tools` | Add tool name (or alias like `search`) to `allowed_tools` |
| Default tool not working | Missing from agent config | Even if `default_tools` is set globally, agent must list the alias in `allowed_tools` |
| Agent ignores skills | Skills not configured | Check `ayo agents show @agent` |
| Agent behaves unsafely | Guardrails disabled | Set `"guardrails": true` |
| Agent not in list | Invalid config | Check config.json is valid JSON |

## Skill Issues

| Problem | Cause | Solution |
|---------|-------|----------|
| Skill not appearing | Name mismatch | Directory name must match `name` field |
| Skill not applied | Not in agent config | Add to `skills` array in config.json |
| Validation fails | Invalid frontmatter | Check YAML syntax in SKILL.md |

## General

- Check `--help` for correct usage
- Run with `--debug` for verbose output
- Verify paths and file permissions
