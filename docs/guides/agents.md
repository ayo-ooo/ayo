# Agent Configuration Guide

Complete reference for configuring ayo agents.

## Directory Structure

```
@agent-name/
├── config.json           # Required: Agent configuration
├── system.md             # Required: System prompt
├── input.jsonschema      # Optional: Input validation
├── output.jsonschema     # Optional: Output format
└── skills/               # Optional: Knowledge files
    └── SKILL.md
```

## config.json Schema

```json
{
  "model": "claude-sonnet-4-20250514",
  "description": "Agent description for ayo agents list",
  "allowed_tools": ["bash", "view", "edit"],
  "disabled_tools": ["delegate"],
  "trust_level": "sandboxed",
  "guardrails": true,
  "skills": ["./skills/SKILL.md"],
  "memory": {
    "enabled": true,
    "scope": "agent",
    "retrieval": {
      "limit": 10,
      "threshold": 0.7
    }
  },
  "permissions": {
    "auto_approve": false,
    "auto_approve_patterns": ["./output/*"]
  },
  "sandbox": {
    "enabled": true,
    "user": "agent",
    "persist_home": true,
    "image": "alpine:latest",
    "network": false,
    "mounts": ["/data:/data:ro"]
  },
  "triggers": [
    {
      "name": "daily-run",
      "type": "cron",
      "schedule": "0 9 * * *",
      "prompt": "Run daily check"
    }
  ],
  "delegates": ["@helper", "@reviewer"]
}
```

## Field Reference

### Core Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | `claude-sonnet-4-20250514` | LLM model to use |
| `description` | string | `""` | Shown in `ayo agents list` |

### Tools

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `allowed_tools` | string[] | All tools | Tools this agent can use |
| `disabled_tools` | string[] | `[]` | Tools explicitly disabled |

**Built-in tools**: `bash`, `view`, `edit`, `glob`, `grep`, `memory_store`, `memory_search`, `file_request`, `publish`, `delegate`, `human_input`

### Trust Levels

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `trust_level` | string | `sandboxed` | Security level |
| `guardrails` | bool | `true` | Enable safety prompts |

**Trust levels**:
- `sandboxed` - Full isolation, read-only host mount
- `privileged` - Sandbox with file_request capability
- `unrestricted` - No sandbox, no guardrails (dangerous)

### Skills

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `skills` | string[] | `[]` | Paths to SKILL.md files |

Paths are relative to the agent directory.

### Memory

```json
{
  "memory": {
    "enabled": true,
    "scope": "agent",
    "formation_triggers": ["preference", "correction"],
    "retrieval": {
      "limit": 10,
      "threshold": 0.7,
      "categories": ["preference", "fact"]
    }
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Enable memory |
| `scope` | string | `agent` | `global`, `agent`, `path`, `squad` |
| `formation_triggers` | string[] | All | Auto-store categories |
| `retrieval.limit` | int | `10` | Max memories per query |
| `retrieval.threshold` | float | `0.7` | Similarity threshold (0-1) |
| `retrieval.categories` | string[] | All | Categories to retrieve |

### Permissions

```json
{
  "permissions": {
    "auto_approve": false,
    "auto_approve_patterns": [
      "./output/*",
      "./tmp/*"
    ]
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `auto_approve` | bool | `false` | Auto-approve all file requests |
| `auto_approve_patterns` | string[] | `[]` | Glob patterns to auto-approve |

### Sandbox

```json
{
  "sandbox": {
    "enabled": true,
    "user": "agent",
    "persist_home": true,
    "image": "alpine:latest",
    "network": false,
    "resources": {
      "memory": "1G",
      "cpu": "1"
    },
    "mounts": [
      "/data:/data:ro",
      "~/docs:/docs:ro"
    ]
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Use sandbox |
| `user` | string | Agent name | Unix user in sandbox |
| `persist_home` | bool | `true` | Persist home directory |
| `image` | string | `alpine:latest` | Container base image |
| `network` | bool | `false` | Allow network access |
| `resources.memory` | string | Unlimited | Memory limit |
| `resources.cpu` | string | Unlimited | CPU limit |
| `mounts` | string[] | `[]` | Additional mounts (source:dest:mode) |

### Triggers

```json
{
  "triggers": [
    {
      "name": "morning-check",
      "type": "cron",
      "schedule": "0 9 * * MON-FRI",
      "prompt": "Run morning checks"
    },
    {
      "name": "file-watch",
      "type": "watch",
      "path": "./src",
      "pattern": "*.go",
      "debounce": "5s",
      "prompt": "Review changed files"
    }
  ]
}
```

### Delegates

```json
{
  "delegates": ["@helper", "@reviewer"]
}
```

Agents this agent can call via the `delegate` tool.

## system.md Format

The system prompt defines agent behavior:

```markdown
# Agent Name

Brief description of what the agent does.

## Responsibilities

- Primary task
- Secondary task

## Guidelines

- How to behave
- What to avoid

## Output Format

How to structure responses.
```

### Best Practices

1. **Be specific**: Vague prompts lead to inconsistent behavior
2. **Include examples**: Show desired output format
3. **Set boundaries**: What should the agent NOT do
4. **Consider context**: Will this run in a squad? With triggers?

## Input/Output Schemas

For agent chaining, define schemas:

**input.jsonschema**:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "code": {
      "type": "string",
      "description": "Code to review"
    },
    "language": {
      "type": "string",
      "enum": ["go", "python", "typescript"]
    }
  },
  "required": ["code"]
}
```

**output.jsonschema**:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "issues": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "severity": { "type": "string" },
          "message": { "type": "string" },
          "line": { "type": "integer" }
        }
      }
    }
  }
}
```

## Agent Locations

Priority order (first match wins):

1. **Project local**: `./.config/ayo/agents/@name/`
2. **User agents**: `~/.config/ayo/agents/@name/`
3. **System agents**: `~/.local/share/ayo/agents/@name/`
4. **Plugin agents**: `~/.local/share/ayo/plugins/*/agents/@name/`

## Complete Examples

### Minimal Agent

```
@simple/
├── config.json
└── system.md
```

**config.json**:
```json
{
  "description": "Simple helper"
}
```

**system.md**:
```markdown
You are a helpful assistant.
```

### Full-Featured Agent

```
@enterprise/
├── config.json
├── system.md
├── input.jsonschema
├── output.jsonschema
└── skills/
    ├── company-guidelines.md
    └── tech-stack.md
```

**config.json**:
```json
{
  "model": "claude-sonnet-4-20250514",
  "description": "Enterprise code assistant",
  "allowed_tools": ["bash", "view", "edit", "grep", "glob"],
  "disabled_tools": ["delegate"],
  "trust_level": "privileged",
  "guardrails": true,
  "skills": [
    "./skills/company-guidelines.md",
    "./skills/tech-stack.md"
  ],
  "memory": {
    "enabled": true,
    "scope": "path",
    "retrieval": {
      "limit": 15,
      "threshold": 0.75
    }
  },
  "permissions": {
    "auto_approve": false,
    "auto_approve_patterns": [
      "./build/*",
      "./dist/*"
    ]
  },
  "sandbox": {
    "persist_home": true,
    "network": true,
    "resources": {
      "memory": "2G"
    }
  }
}
```

## Troubleshooting

### Agent not loading

```bash
# Check syntax
cat ~/.config/ayo/agents/@name/config.json | jq .

# Verify directory exists
ls ~/.config/ayo/agents/@name/
```

### Wrong model being used

Resolution order:
1. Agent config.json `model`
2. Global config `model`
3. `AYO_MODEL` environment variable
4. Default

### Tools not available

Check `allowed_tools` includes what you need, and `disabled_tools` doesn't exclude it.

### Memory not working

Ensure embedding provider is configured:
```bash
ayo doctor | grep embedding
```
