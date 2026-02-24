# Plugin Reference

Complete reference for ayo plugin development and distribution.

## Overview

Plugins extend ayo with additional agents, tools, skills, providers, planners, squads, and triggers. They are distributed as Git repositories or local directories.

## Plugin Structure

```
my-plugin/
├── manifest.json          # Required: plugin metadata
├── README.md              # Recommended: documentation
├── LICENSE                # Recommended: license file
├── agents/                # Agent definitions
│   └── my-agent/
│       ├── agent.md       # Agent system prompt
│       └── ayo.json       # Agent configuration
├── skills/                # Skill files
│   └── my-skill/
│       └── skill.md
├── tools/                 # External tools
│   └── my-tool/
│       ├── tool.json      # Tool definition
│       └── run.sh         # Tool implementation
├── flows/                 # Flow definitions
│   └── my-flow.yaml
└── scripts/               # Support scripts
    └── post-install.sh
```

## Manifest Schema

The `manifest.json` file defines plugin metadata and components.

### Required Fields

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "A useful plugin for ayo"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Plugin identifier (lowercase, hyphens allowed) |
| `version` | string | Semantic version (e.g., "1.0.0") |
| `description` | string | Brief description of the plugin |

### Optional Metadata

```json
{
  "author": "Your Name <email@example.com>",
  "repository": "https://github.com/user/ayo-plugins-example",
  "license": "Apache-2.0",
  "ayo_version": ">=0.2.0"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `author` | string | Plugin author name and email |
| `repository` | string | Git repository URL |
| `license` | string | SPDX license identifier |
| `ayo_version` | string | Required ayo version constraint |

### Component Arrays

#### Agents

```json
{
  "agents": ["my-agent", "another-agent"]
}
```

Lists agent handles. Each must have a corresponding `agents/<name>/` directory with `agent.md` and optionally `ayo.json`.

#### Skills

```json
{
  "skills": ["my-skill"]
}
```

Lists skill names. Each must have a corresponding `skills/<name>/` directory.

#### Tools

```json
{
  "tools": ["my-tool"]
}
```

Lists tool names. Each must have a `tools/<name>/tool.json` definition.

### Delegates

Map task types to agents that handle them:

```json
{
  "delegates": {
    "code-review": "@reviewer",
    "testing": "@tester"
  }
}
```

### Default Tools

Override default tool implementations:

```json
{
  "default_tools": {
    "editor": "my-editor-tool",
    "terminal": "my-terminal-tool"
  }
}
```

### Dependencies

```json
{
  "dependencies": {
    "binaries": [
      "git",
      {
        "name": "rg",
        "install_hint": "ripgrep",
        "install_url": "https://github.com/BurntSushi/ripgrep",
        "install_cmd": "brew install ripgrep"
      }
    ],
    "plugins": ["other-plugin"]
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `binaries` | array | Required system binaries |
| `plugins` | array | Required ayo plugins |

Binary entries can be strings (simple name) or objects with installation hints.

### Post-Install Script

```json
{
  "post_install": "scripts/post-install.sh"
}
```

Script executed after plugin installation. Must be executable.

### Providers

```json
{
  "providers": [
    {
      "name": "my-memory",
      "type": "memory",
      "entry_point": "providers/memory.so",
      "config": {}
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Provider identifier |
| `type` | string | One of: `memory`, `sandbox`, `embedding`, `observer` |
| `entry_point` | string | Path to Go plugin (.so) or executable |
| `config` | object | Provider-specific configuration |

**Provider Types:**

| Type | Description |
|------|-------------|
| `memory` | Memory storage backend |
| `sandbox` | Sandbox execution environment |
| `embedding` | Text embedding service |
| `observer` | Agent activity observer |

### Planners

```json
{
  "planners": [
    {
      "name": "my-planner",
      "type": "near",
      "entry_point": "planners/near.so",
      "config": {}
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Planner identifier |
| `type` | string | `near` (session-scoped) or `long` (persistent) |
| `entry_point` | string | Path to Go plugin (.so) |
| `config` | object | Planner-specific configuration |

**Planner Types:**

| Type | Description | Example |
|------|-------------|---------|
| `near` | Session-scoped task tracking | Todo lists, in-memory tasks |
| `long` | Persistent coordination | Ticket systems, project management |

### Sandbox Configs

Pre-configured sandbox environments:

```json
{
  "sandbox_configs": [
    {
      "name": "node-dev",
      "description": "Node.js development environment",
      "image": "node:20",
      "packages": ["git", "vim"],
      "env": {
        "NODE_ENV": "development"
      }
    }
  ]
}
```

### Squads

Multi-agent team definitions:

```json
{
  "squads": [
    {
      "name": "dev-team",
      "description": "Full development team",
      "path": "squads/dev-team",
      "agents": ["code", "reviewer", "tester"],
      "planners": {
        "near": "ayo-todos",
        "long": "ayo-tickets"
      }
    }
  ]
}
```

### Triggers

Event-based agent activation:

```json
{
  "triggers": [
    {
      "name": "file-watcher",
      "category": "watch",
      "entry_point": "triggers/watcher.so",
      "config_schema": {
        "type": "object",
        "properties": {
          "path": {"type": "string"},
          "pattern": {"type": "string"}
        }
      }
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Trigger identifier |
| `category` | string | `poll`, `push`, or `watch` |
| `entry_point` | string | Path to Go plugin (.so) |
| `config_schema` | object | JSON Schema for trigger configuration |

**Trigger Categories:**

| Category | Description |
|----------|-------------|
| `poll` | Periodic polling (cron-like) |
| `push` | External webhook/event |
| `watch` | File system monitoring |

## Complete Example

```json
{
  "name": "ayo-plugins-devtools",
  "version": "2.1.0",
  "description": "Development tools and agents for ayo",
  "author": "Ayo Team <team@ayo.dev>",
  "repository": "https://github.com/anthropics/ayo-plugins-devtools",
  "license": "Apache-2.0",
  "ayo_version": ">=0.3.0",

  "agents": ["code", "reviewer", "tester"],
  "skills": ["golang", "typescript", "testing"],
  "tools": ["lint", "format", "test-runner"],

  "delegates": {
    "code-review": "@reviewer",
    "testing": "@tester",
    "debugging": "@code"
  },

  "dependencies": {
    "binaries": [
      "git",
      {"name": "golangci-lint", "install_cmd": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"}
    ]
  },

  "planners": [
    {
      "name": "sprint-planner",
      "type": "long",
      "entry_point": "planners/sprint.so"
    }
  ],

  "squads": [
    {
      "name": "full-stack",
      "description": "Complete development team",
      "agents": ["code", "reviewer", "tester"],
      "planners": {"long": "sprint-planner"}
    }
  ],

  "triggers": [
    {
      "name": "pr-review",
      "category": "push",
      "entry_point": "triggers/github-pr.so"
    }
  ],

  "post_install": "scripts/setup.sh"
}
```

## Tool Definition

Tools are defined in `tools/<name>/tool.json`:

```json
{
  "name": "my-tool",
  "description": "Does something useful",
  "parameters": {
    "type": "object",
    "properties": {
      "input": {
        "type": "string",
        "description": "Input value"
      }
    },
    "required": ["input"]
  },
  "command": ["./run.sh", "{{input}}"],
  "working_dir": "{{sandbox_cwd}}",
  "timeout": 30
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Tool name |
| `description` | string | Tool description for LLM |
| `parameters` | object | JSON Schema for parameters |
| `command` | array | Command to execute |
| `working_dir` | string | Working directory |
| `timeout` | number | Timeout in seconds |

### Template Variables

| Variable | Description |
|----------|-------------|
| `{{param}}` | Parameter value |
| `{{sandbox_cwd}}` | Current sandbox working directory |
| `{{sandbox_home}}` | Sandbox home directory |
| `{{plugin_dir}}` | Plugin installation directory |

## Installation

### From Git Repository

```bash
ayo plugin install https://github.com/user/ayo-plugins-example
```

Repository names should follow the convention `ayo-plugins-<name>`.

### From Local Directory

```bash
ayo plugin install ~/path/to/plugin
```

### Listing Installed Plugins

```bash
ayo plugin list
```

### Removing Plugins

```bash
ayo plugin remove my-plugin
```

## Plugin Development

### Creating a New Plugin

1. Create directory structure:
   ```bash
   mkdir -p my-plugin/{agents,skills,tools}
   ```

2. Create `manifest.json`:
   ```bash
   cat > my-plugin/manifest.json << 'EOF'
   {
     "name": "my-plugin",
     "version": "0.1.0",
     "description": "My first plugin"
   }
   EOF
   ```

3. Add components as needed

4. Test locally:
   ```bash
   ayo plugin install ~/my-plugin
   ```

### Go Plugin Interface

For providers, planners, and triggers, implement the appropriate interface:

**Planner Interface** (`internal/planners/interface.go`):

```go
type PlannerPlugin interface {
    // Initialize the planner
    Init(config map[string]any) error
    
    // Get planner type
    Type() PlannerType  // NearTerm or LongTerm
    
    // Get tools provided by this planner
    Tools() []Tool
    
    // Get system prompt injection
    SystemPrompt() string
    
    // Cleanup resources
    Close() error
}
```

Build as a Go plugin:

```bash
go build -buildmode=plugin -o my-planner.so planner.go
```

### Testing Plugins

```bash
# Validate manifest
ayo plugin validate ~/my-plugin

# Install and test
ayo plugin install ~/my-plugin
ayo run my-agent "Test the agent"
```

## Distribution

### Git Repository

1. Create repository named `ayo-plugins-<name>`
2. Tag releases with semantic versions: `git tag v1.0.0`
3. Users install via: `ayo plugin install https://github.com/user/ayo-plugins-<name>`

### Registry (Coming Soon)

```bash
# Future: publish to registry
ayo plugin publish

# Future: install from registry
ayo plugin install <name>
```

## Plugin Locations

| Location | Description |
|----------|-------------|
| `~/.config/ayo/plugins/` | Installed plugins |
| `~/.local/share/ayo/packages.json` | Plugin registry |

## See Also

- [Guides: Tools](../guides/tools.md) - Tool system overview
- [Concepts](../concepts.md) - Core ayo concepts
- [Tutorials: Plugins](../tutorials/plugins.md) - Creating your first plugin
