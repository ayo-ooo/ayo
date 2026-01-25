# Ayo Plugin System

Plugins extend ayo with additional agents, skills, and tools. They're distributed via git repositories and installed with a single command.

## Table of Contents

- [Overview](#overview)
- [Installing Plugins](#installing-plugins)
- [Managing Plugins](#managing-plugins)
- [Creating Plugins](#creating-plugins)
  - [Repository Structure](#repository-structure)
  - [The Manifest File](#the-manifest-file)
  - [Adding Agents](#adding-agents)
  - [Adding Skills](#adding-skills)
  - [Adding Tools](#adding-tools)
  - [Declaring Delegates](#declaring-delegates)
- [Tool Definition Reference](#tool-definition-reference)
- [Examples](#examples)
  - [Simple Tool Plugin](#example-1-simple-tool-plugin)
  - [Agent with Custom Skill](#example-2-agent-with-custom-skill)
  - [Full-Featured Plugin](#example-3-full-featured-plugin)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

A plugin is a git repository containing:

| Component | Directory | Description |
|-----------|-----------|-------------|
| Manifest | `manifest.json` | Required metadata about the plugin |
| Agents | `agents/` | AI agents with custom prompts and tools |
| Skills | `skills/` | Domain-specific instructions for agents |
| Tools | `tools/` | External CLI commands exposed to agents |

When installed, plugins are cloned to `~/.local/share/ayo/plugins/<name>/` and registered in `~/.local/share/ayo/packages.json`.

## Installing Plugins

### From Git URL

```bash
# HTTPS
ayo plugins install https://github.com/user/ayo-plugins-example

# SSH
ayo plugins install git@github.com:user/ayo-plugins-example.git

# GitLab, Bitbucket, or any git host
ayo plugins install https://gitlab.com/org/ayo-plugins-tools.git
```

### Repository Naming Convention

Plugin repositories should follow the naming pattern `ayo-plugins-<name>`:

- `ayo-plugins-crush` → plugin name: `crush`
- `ayo-plugins-research` → plugin name: `research`
- `ayo-plugins-my-tools` → plugin name: `my-tools`

If the repository doesn't follow this convention, the full repo name is used as the plugin name.

### Installation Options

```bash
# Force reinstall (overwrites existing)
ayo plugins install https://github.com/user/repo --force

# Install from local directory (for development)
ayo plugins install --local ./my-plugin

# Skip delegate configuration prompt
ayo plugins install https://github.com/user/repo --yes
```

### Dependency Checking

During installation, ayo checks for required dependencies and offers to install them:

```
✓ Installed crush v1.0.0
  → Agents: @crush
  → Skills: crush-coding
  → Tools: crush

! Missing dependencies:
  ✗ crush
    Install with: go install github.com/charmbracelet/crush@latest

? Install crush now? [Y/n]
```

If a dependency has an `install_cmd` defined, you'll be prompted to install it automatically.
Otherwise, the install hint and/or URL will be displayed so you can install manually.

## Managing Plugins

### List Installed Plugins

```bash
ayo plugins list
```

Output:
```
Installed Plugins
─────────────────────────────────────────────

◆ crush v1.0.0
  Crush coding agent for ayo
  Agents: @crush
  Skills: crush-coding
  Tools: crush
```

### Show Plugin Details

```bash
ayo plugins show crush
```

### Update Plugins

```bash
# Update all plugins
ayo plugins update

# Update specific plugin
ayo plugins update crush

# Force update (ignore local changes)
ayo plugins update crush --force

# Check for updates without applying
ayo plugins update --dry-run
```

### Remove Plugins

```bash
# Remove with confirmation prompt
ayo plugins remove crush

# Remove without confirmation
ayo plugins remove crush --yes
```

## Creating Plugins

### Repository Structure

```
ayo-plugins-<name>/
├── manifest.json           # Required: plugin metadata
├── README.md               # Recommended: documentation
├── agents/                 # Optional: agent definitions
│   └── @agent-name/
│       ├── config.json     # Agent configuration
│       ├── system.md       # System prompt
│       └── skills/         # Agent-specific skills
│           └── my-skill/
│               └── SKILL.md
├── skills/                 # Optional: shared skills
│   └── skill-name/
│       └── SKILL.md
└── tools/                  # Optional: external tools
    └── tool-name/
        └── tool.json
```

### The Manifest File

Every plugin requires a `manifest.json` at the repository root:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "A brief description of what this plugin provides",
  "author": "your-name",
  "repository": "https://github.com/user/ayo-plugins-my-plugin",
  "license": "MIT",
  "agents": ["@my-agent"],
  "skills": ["my-skill"],
  "tools": ["my-tool"],
  "delegates": {
    "coding": "@my-agent"
  },
  "dependencies": {
    "binaries": [
      "simple-dep",
      {
        "name": "crush",
        "install_hint": "Install with: go install github.com/charmbracelet/crush@latest",
        "install_cmd": "go install github.com/charmbracelet/crush@latest",
        "install_url": "https://github.com/charmbracelet/crush"
      }
    ]
  },
  "ayo_version": ">=0.2.0"
}
```

#### Required Fields

| Field | Description |
|-------|-------------|
| `name` | Plugin identifier. Lowercase alphanumeric with hyphens. |
| `version` | Semantic version (e.g., `1.0.0`, `0.2.1-beta`). |
| `description` | Brief description of the plugin. |

#### Optional Fields

| Field | Description |
|-------|-------------|
| `author` | Author name or organization. |
| `repository` | Git repository URL. |
| `license` | SPDX license identifier (e.g., `MIT`, `Apache-2.0`). |
| `agents` | List of agent handles provided (must exist in `agents/`). |
| `skills` | List of skill names provided (must exist in `skills/`). |
| `tools` | List of tool names provided (must exist in `tools/`). |
| `delegates` | Task types this plugin handles (see [Delegates](#declaring-delegates)). |
| `dependencies` | External requirements (see [Dependencies](#dependencies)). |
| `ayo_version` | Minimum ayo version required (semver constraint). |

#### Dependencies

The `dependencies` field specifies external requirements. Binary dependencies can be simple strings or objects with installation instructions:

```json
{
  "dependencies": {
    "binaries": [
      "simple-binary",
      {
        "name": "complex-binary",
        "install_hint": "Human-readable installation instructions",
        "install_cmd": "go install github.com/example/tool@latest",
        "install_url": "https://example.com/install"
      }
    ],
    "plugins": ["other-ayo-plugin"]
  }
}
```

| Binary Field | Description |
|--------------|-------------|
| `name` | Required. The binary name to look for in PATH. |
| `install_hint` | Human-readable message shown when dependency is missing. |
| `install_cmd` | Command to run to install the dependency. If provided, user is prompted to run it. |
| `install_url` | URL with installation instructions. |

### Adding Agents

Agents live in `agents/@<handle>/`:

```
agents/
└── @my-agent/
    ├── config.json
    ├── system.md
    └── skills/           # Optional agent-specific skills
```

#### Agent config.json

```json
{
  "description": "What this agent does",
  "model": "gpt-4.1",
  "allowed_tools": ["bash", "my-tool"],
  "skills": ["my-skill"],
  "exclude_skills": ["unwanted-skill"],
  "guardrails": true
}
```

| Field | Description |
|-------|-------------|
| `description` | Brief agent description. |
| `model` | Default model (optional, uses global default if omitted). |
| `allowed_tools` | Tools the agent can use. |
| `skills` | Skills to attach (in addition to auto-discovered). |
| `exclude_skills` | Skills to exclude from auto-discovery. |
| `guardrails` | Safety guardrails (default: true). Set to false to disable (dangerous). |

#### Agent system.md

The system prompt that defines the agent's behavior:

```markdown
You are a specialized assistant for [task domain].

## Guidelines

1. Be concise and accurate
2. Focus on [specific outcomes]
3. Use tools when needed

## Capabilities

- [Capability 1]
- [Capability 2]
```

### Adding Skills

Skills live in `skills/<name>/`:

```
skills/
└── my-skill/
    ├── SKILL.md          # Required
    ├── scripts/          # Optional executables
    ├── references/       # Optional documentation
    └── assets/           # Optional data files
```

#### SKILL.md Format

```markdown
---
name: my-skill
description: |
  What this skill does and when to use it.
  This is shown to the agent to help it decide when to apply the skill.
metadata:
  author: your-name
  version: "1.0"
---

# Skill Instructions

Detailed instructions for the agent on how to use this skill.

## When to Use

- Scenario 1
- Scenario 2

## How to Use

Step-by-step guidance...
```

The frontmatter fields:

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Skill identifier (must match directory name). |
| `description` | Yes | When to use this skill (1-1024 chars). |
| `metadata` | No | Key-value pairs (author, version, etc.). |
| `compatibility` | No | Environment requirements (max 500 chars). |

### Adding Tools

Tools wrap external CLI commands. They live in `tools/<name>/`:

```
tools/
└── my-tool/
    └── tool.json
```

See [Tool Definition Reference](#tool-definition-reference) for the full specification.

### Declaring Delegates

Delegates let plugins handle specific task types. When a plugin declares delegates, users are prompted during installation to set them as global defaults.

```json
{
  "delegates": {
    "coding": "@my-coding-agent",
    "research": "@my-research-agent"
  }
}
```

Supported task types:

| Type | Description |
|------|-------------|
| `coding` | Source code creation and modification |
| `research` | Web research and information gathering |
| `debug` | Debugging and troubleshooting |
| `test` | Test creation and execution |
| `docs` | Documentation generation |

When a delegate is configured, `@ayo` will automatically route tasks of that type to the delegate agent.

## Tool Definition Reference

Tools are defined in `tools/<name>/tool.json`:

```json
{
  "name": "my-tool",
  "description": "What this tool does",
  "command": "my-binary",
  "args": ["--flag", "value"],
  "parameters": [...],
  "timeout": 60,
  "working_dir": "inherit",
  "env": {
    "MY_VAR": "value"
  },
  "depends_on": ["my-binary"]
}
```

### Tool Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Tool identifier. |
| `description` | Yes | What the tool does (shown to LLM). |
| `command` | Yes | Executable to run (binary name or path). |
| `args` | No | Default arguments passed to command. |
| `parameters` | No | Input parameters for the tool. |
| `timeout` | No | Timeout in seconds (0 = no timeout). |
| `working_dir` | No | `inherit`, `plugin`, or `param`. |
| `allow_any_dir` | No | Allow any directory for working_dir param. |
| `quiet` | No | Suppress output in UI. |
| `stream_output` | No | Stream output as produced. |
| `env` | No | Environment variables to set. |
| `depends_on` | No | Required binaries. |

### Parameter Definition

```json
{
  "name": "input",
  "description": "The input to process",
  "type": "string",
  "required": true,
  "default": null,
  "enum": ["option1", "option2"],
  "arg_template": "--input={{value}}",
  "position": 0,
  "omit_if_empty": true
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Parameter identifier. |
| `description` | Yes | What the parameter does. |
| `type` | Yes | `string`, `number`, `integer`, `boolean`, `array`, `object`. |
| `required` | No | Whether parameter must be provided. |
| `default` | No | Default value if not provided. |
| `enum` | No | Allowed values for string type. |
| `arg_template` | No | How to map to command args (e.g., `--flag={{value}}`). |
| `position` | No | Position for positional arguments (0-based). |
| `omit_if_empty` | No | Skip arg if value is empty/false/nil. |

### Argument Templates

The `arg_template` field controls how parameters become command arguments:

```json
// Template: "--model={{value}}"
// Input: {"model": "gpt-4"}
// Result: --model=gpt-4

// Template: "-f {{value}}"
// Input: {"file": "data.json"}
// Result: -f data.json

// Template: "--verbose" (boolean flag)
// Input: {"verbose": true}
// Result: --verbose

// No template (uses default)
// Input: {"name": "value"}
// Result: --name=value
```

For positional arguments, set `position` and use `{{value}}`:

```json
{
  "name": "prompt",
  "position": 0,
  "arg_template": "{{value}}"
}
// Input: {"prompt": "hello world"}
// Result: "hello world" as first positional arg
```

## Examples

### Example 1: Simple Tool Plugin

A plugin that wraps a CLI tool:

**Repository: `ayo-plugins-jq`**

```
ayo-plugins-jq/
├── manifest.json
└── tools/
    └── jq/
        └── tool.json
```

**manifest.json:**
```json
{
  "name": "jq",
  "version": "1.0.0",
  "description": "JQ JSON processor for ayo agents",
  "tools": ["jq"],
  "dependencies": {
    "binaries": ["jq"]
  }
}
```

**tools/jq/tool.json:**
```json
{
  "name": "jq",
  "description": "Process JSON with jq expressions. Use for filtering, transforming, and querying JSON data.",
  "command": "jq",
  "parameters": [
    {
      "name": "filter",
      "description": "The jq filter expression (e.g., '.foo', '.[] | select(.active)')",
      "type": "string",
      "required": true,
      "position": 0
    },
    {
      "name": "input",
      "description": "JSON input to process",
      "type": "string",
      "required": true,
      "arg_template": "{{value}}",
      "position": 1
    },
    {
      "name": "raw",
      "description": "Output raw strings without quotes",
      "type": "boolean",
      "required": false,
      "arg_template": "-r",
      "omit_if_empty": true
    }
  ],
  "timeout": 30
}
```

**Usage after installation:**
```bash
ayo plugins install https://github.com/user/ayo-plugins-jq

# Agent can now use jq tool
ayo @ayo "parse this JSON and extract all names: {\"users\":[{\"name\":\"Alice\"},{\"name\":\"Bob\"}]}"
```

### Example 2: Agent with Custom Skill

A specialized agent with its own skill:

**Repository: `ayo-plugins-sql`**

```
ayo-plugins-sql/
├── manifest.json
├── agents/
│   └── @sql/
│       ├── config.json
│       └── system.md
└── skills/
    └── sql-expert/
        └── SKILL.md
```

**manifest.json:**
```json
{
  "name": "sql",
  "version": "1.0.0",
  "description": "SQL query expert agent",
  "agents": ["@sql"],
  "skills": ["sql-expert"]
}
```

**agents/@sql/config.json:**
```json
{
  "description": "SQL expert for writing and optimizing queries",
  "allowed_tools": ["bash"],
  "skills": ["sql-expert"]
}
```

**agents/@sql/system.md:**
```markdown
You are an expert SQL developer. You help users write, debug, and optimize SQL queries.

## Capabilities

- Write complex SQL queries (joins, subqueries, CTEs, window functions)
- Optimize slow queries with EXPLAIN analysis
- Convert between SQL dialects (PostgreSQL, MySQL, SQLite)
- Design database schemas

## Guidelines

1. Always ask which database system (PostgreSQL, MySQL, SQLite) if not specified
2. Explain query logic briefly
3. Suggest indexes for performance when relevant
4. Use CTEs for readability over deeply nested subqueries
```

**skills/sql-expert/SKILL.md:**
```markdown
---
name: sql-expert
description: |
  SQL query writing and optimization expertise.
  Use when working with databases, writing queries, or analyzing schemas.
metadata:
  author: sql-plugin
  version: "1.0"
---

# SQL Expert Skill

## Query Writing Patterns

### Common Table Expressions (CTEs)
Use CTEs for complex queries:
```sql
WITH active_users AS (
  SELECT * FROM users WHERE status = 'active'
)
SELECT * FROM active_users WHERE created_at > '2024-01-01';
```

### Window Functions
For ranking and aggregation:
```sql
SELECT 
  name,
  sales,
  RANK() OVER (ORDER BY sales DESC) as rank
FROM salespeople;
```

## Optimization Tips

1. Check EXPLAIN output for sequential scans
2. Add indexes on frequently filtered columns
3. Use LIMIT when sampling data
4. Avoid SELECT * in production queries
```

**Usage:**
```bash
ayo plugins install https://github.com/user/ayo-plugins-sql
ayo @sql "write a query to find the top 10 customers by order value"
```

### Example 3: Full-Featured Plugin

A complete plugin with agent, skill, tool, and delegation:

**Repository: `ayo-plugins-docker`**

```
ayo-plugins-docker/
├── manifest.json
├── README.md
├── agents/
│   └── @docker/
│       ├── config.json
│       └── system.md
├── skills/
│   └── docker-ops/
│       └── SKILL.md
└── tools/
    └── docker-compose/
        └── tool.json
```

**manifest.json:**
```json
{
  "name": "docker",
  "version": "1.0.0",
  "description": "Docker container management agent with compose support",
  "author": "devops-team",
  "repository": "https://github.com/org/ayo-plugins-docker",
  "license": "MIT",
  "agents": ["@docker"],
  "skills": ["docker-ops"],
  "tools": ["docker-compose"],
  "dependencies": {
    "binaries": ["docker", "docker-compose"]
  },
  "ayo_version": ">=0.2.0"
}
```

**agents/@docker/config.json:**
```json
{
  "description": "Docker container and compose management",
  "allowed_tools": ["bash", "docker-compose"],
  "skills": ["docker-ops"]
}
```

**agents/@docker/system.md:**
```markdown
You are a Docker operations expert. You help manage containers, images, and docker-compose configurations.

## Capabilities

- Build and manage Docker images
- Run and orchestrate containers
- Write and debug docker-compose files
- Troubleshoot container networking
- Optimize Dockerfiles for size and build speed

## Guidelines

1. Always check if containers are running before operations
2. Use docker-compose for multi-container setups
3. Prefer named volumes over bind mounts for data persistence
4. Include health checks in production configurations
```

**skills/docker-ops/SKILL.md:**
```markdown
---
name: docker-ops
description: |
  Docker operations including container management, image building,
  and docker-compose orchestration.
metadata:
  version: "1.0"
---

# Docker Operations

## Common Commands

### Container Management
```bash
docker ps                    # List running containers
docker ps -a                 # List all containers
docker logs <container>      # View logs
docker exec -it <c> bash     # Shell into container
```

### Image Management
```bash
docker images                # List images
docker build -t name:tag .   # Build image
docker push name:tag         # Push to registry
```

## Compose Patterns

### Development Setup
```yaml
services:
  app:
    build: .
    volumes:
      - .:/app
    environment:
      - DEBUG=true
```

### Production Setup
```yaml
services:
  app:
    image: myapp:latest
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/health"]
```
```

**tools/docker-compose/tool.json:**
```json
{
  "name": "docker-compose",
  "description": "Run docker-compose commands for multi-container Docker applications",
  "command": "docker-compose",
  "parameters": [
    {
      "name": "action",
      "description": "The compose action: up, down, build, logs, ps, restart",
      "type": "string",
      "required": true,
      "enum": ["up", "down", "build", "logs", "ps", "restart", "exec"],
      "position": 0
    },
    {
      "name": "service",
      "description": "Target service name (optional, applies to all if omitted)",
      "type": "string",
      "required": false,
      "position": 1,
      "omit_if_empty": true
    },
    {
      "name": "detach",
      "description": "Run in detached mode (for 'up' action)",
      "type": "boolean",
      "required": false,
      "arg_template": "-d",
      "omit_if_empty": true
    },
    {
      "name": "build",
      "description": "Build images before starting (for 'up' action)",
      "type": "boolean",
      "required": false,
      "arg_template": "--build",
      "omit_if_empty": true
    },
    {
      "name": "file",
      "description": "Path to docker-compose file",
      "type": "string",
      "required": false,
      "arg_template": "-f {{value}}",
      "omit_if_empty": true
    }
  ],
  "timeout": 300,
  "depends_on": ["docker-compose"]
}
```

**Usage:**
```bash
ayo plugins install https://github.com/org/ayo-plugins-docker
ayo @docker "start the development environment"
ayo @docker "show me logs for the api service"
```

## Best Practices

### Plugin Design

1. **Single Responsibility**: Each plugin should focus on one domain or tool
2. **Clear Documentation**: Include a README with installation and usage examples
3. **Semantic Versioning**: Use semver for predictable updates
4. **Dependency Declaration**: Always declare required binaries

### Agent Design

1. **Focused Prompts**: Keep system prompts concise and action-oriented
2. **Appropriate Tools**: Only include tools the agent actually needs
3. **Skill Integration**: Use skills for reusable domain knowledge

### Tool Design

1. **Descriptive Names**: Tool name should indicate its purpose
2. **Clear Parameters**: Each parameter needs a helpful description
3. **Sensible Defaults**: Provide defaults where appropriate
4. **Timeout Settings**: Set reasonable timeouts for long-running commands

### Security

1. **No Secrets in Repos**: Never commit API keys or credentials
2. **Validate Inputs**: Be cautious with user-provided paths
3. **Limit Scope**: Use `working_dir: "param"` carefully
4. **Document Permissions**: Note any elevated permissions needed

## Troubleshooting

### Plugin Won't Install

```
Error: manifest: name is required
```
Check that `manifest.json` exists and has all required fields.

```
Error: manifest: declared agent not found in agents/ directory
```
Verify the agent directory matches the handle in `agents` array (including the `@` prefix).

### Missing Dependencies

```
✗ Missing binary: some-tool
```
Install the required binary before using the plugin.

### Agent Not Found After Install

```bash
ayo agents list  # Check if agent appears
ayo plugins show <name>  # Verify installation
```

If the agent doesn't appear, try:
```bash
ayo plugins remove <name> --yes
ayo plugins install <url> --force
```

### Tool Not Working

1. Test the command manually in your terminal
2. Check `ayo plugins show <name>` for tool details
3. Verify binary is in PATH: `which <command>`
4. Check tool.json parameter definitions
