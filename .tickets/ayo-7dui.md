---
id: ayo-7dui
status: open
deps: []
links: []
created: 2026-02-23T23:13:19Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-pv3a
tags: [schema, agents]
---
# Design ayo.json schema with agent namespace

Create unified ayo.json schema for agent configuration.

## Schema Design

```json
{
  "$schema": "https://ayo.dev/schemas/ayo.json",
  "version": "1",
  
  "agent": {
    "description": "A helpful coding assistant",
    "model": "claude-sonnet-4-5-20250929",
    "model_config": {
      "temperature": 0.7,
      "max_tokens": 4096
    },
    "tools": ["bash", "memory", "file_request"],
    "tools_disabled": ["web_search"],
    "skills": ["coding", "debugging"],
    "memory": {
      "enabled": true,
      "scope": "global"
    },
    "sandbox": {
      "isolated": false,
      "network": true,
      "image": "alpine:3.21"
    },
    "permissions": {
      "auto_approve": false
    },
    "delegates": ["@reviewer", "@tester"],
    "triggers": [
      {
        "name": "on-push",
        "type": "watch",
        "pattern": "*.go"
      }
    ]
  }
}
```

## Field Definitions

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Human-readable agent description |
| `model` | string | LLM model identifier |
| `model_config` | object | Model-specific settings |
| `tools` | string[] | Enabled tools (default: all) |
| `tools_disabled` | string[] | Explicitly disabled tools |
| `skills` | string[] | Agent skills to load |
| `memory.enabled` | bool | Enable persistent memory |
| `memory.scope` | string | "global", "agent", or "session" |
| `sandbox.isolated` | bool | Run in own sandbox (vs @ayo) |
| `sandbox.network` | bool | Allow network access |
| `sandbox.image` | string | Container base image |
| `permissions.auto_approve` | bool | Auto-approve file_request |
| `delegates` | string[] | Agents this agent can invoke |
| `triggers` | object[] | Trigger definitions |

## Go Struct Updates

Update `internal/agent/agent.go`:

```go
type AyoConfig struct {
    Schema  string       `json:"$schema,omitempty"`
    Version string       `json:"version"`
    Agent   *AgentConfig `json:"agent,omitempty"`
    Squad   *SquadConfig `json:"squad,omitempty"`
}

type AgentConfig struct {
    Description   string            `json:"description,omitempty"`
    Model         string            `json:"model,omitempty"`
    ModelConfig   *ModelConfig      `json:"model_config,omitempty"`
    Tools         []string          `json:"tools,omitempty"`
    ToolsDisabled []string          `json:"tools_disabled,omitempty"`
    Skills        []string          `json:"skills,omitempty"`
    Memory        *MemoryConfig     `json:"memory,omitempty"`
    Sandbox       *SandboxConfig    `json:"sandbox,omitempty"`
    Permissions   *PermissionsConfig `json:"permissions,omitempty"`
    Delegates     []string          `json:"delegates,omitempty"`
    Triggers      []TriggerConfig   `json:"triggers,omitempty"`
}
```

## Files to Create/Modify

1. Create `schemas/ayo.json` - JSON Schema file
2. Update `internal/agent/agent.go` - Go struct definitions
3. Create `internal/config/ayo_config.go` - Loader functions

## Testing

- Validate schema with various configs
- Test backwards compatibility with config.json
- Test schema validation errors have clear messages
