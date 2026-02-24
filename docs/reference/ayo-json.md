# ayo.json Reference

Complete schema reference for ayo.json configuration files.

## Overview

`ayo.json` (or `config.json` for agents) configures agents, squads, and global settings.

## Locations

| Location | Purpose |
|----------|---------|
| `~/.config/ayo/config.json` | Global defaults |
| `~/.config/ayo/agents/@name/config.json` | Agent config |
| `~/.local/share/ayo/sandboxes/squads/{name}/ayo.json` | Squad config |
| `./.config/ayo/ayo.json` | Project-level config |

---

## Global Configuration

`~/.config/ayo/config.json`:

```json
{

  "provider": "your-provider",
  "model": "your-model",
  "sandbox": {
    "provider": "applecontainer",
    "default_image": "alpine:latest"
  },
  "permissions": {
    "no_jodas": false,
    "blocked_patterns": []
  },
  "memory": {
    "enabled": true,
    "embedding_provider": "anthropic"
  },
  "audit": {
    "enabled": true,
    "retention_days": 90
  }
}
```

### Global Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `provider` | string | (none) | LLM provider |
| `model` | string | (none) | Default model |
| `sandbox` | object | | Sandbox defaults |
| `permissions` | object | | Permission settings |
| `memory` | object | | Memory settings |
| `audit` | object | | Audit settings |

---

## Agent Configuration

`~/.config/ayo/agents/@name/config.json`:

```json
{

  "model": "your-model",
  "description": "Agent description",
  "allowed_tools": ["bash", "view", "edit"],
  "disabled_tools": [],
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
    "auto_approve_patterns": []
  },
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
    "mounts": []
  },
  "triggers": [],
  "delegates": []
}
```

### Agent Fields

#### Core

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | (global) | LLM model |
| `description` | string | `""` | Agent description |

#### Tools

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `allowed_tools` | string[] | All | Enabled tools |
| `disabled_tools` | string[] | `[]` | Disabled tools |

#### Security

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `trust_level` | enum | `sandboxed` | `sandboxed`, `privileged`, `unrestricted` |
| `guardrails` | bool | `true` | Enable safety prompts |

#### Skills

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `skills` | string[] | `[]` | Paths to SKILL.md files |

#### Memory

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
| `scope` | enum | `agent` | `global`, `agent`, `path`, `squad` |
| `formation_triggers` | string[] | All | Auto-store categories |
| `retrieval.limit` | int | `10` | Max memories per query |
| `retrieval.threshold` | float | `0.7` | Similarity threshold (0-1) |
| `retrieval.categories` | string[] | All | Categories to retrieve |

#### Permissions

```json
{
  "permissions": {
    "auto_approve": false,
    "auto_approve_patterns": ["./build/*", "./output/*"]
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `auto_approve` | bool | `false` | Auto-approve all requests |
| `auto_approve_patterns` | string[] | `[]` | Patterns to auto-approve |

#### Sandbox

```json
{
  "sandbox": {
    "enabled": true,
    "user": "agent",
    "persist_home": true,
    "image": "alpine:latest",
    "network": false,
    "resources": {
      "memory": "2G",
      "cpu": "2",
      "disk": "10G"
    },
    "mounts": ["/data:/data:ro"]
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Use sandbox |
| `user` | string | Agent name | Unix user |
| `persist_home` | bool | `true` | Persist home dir |
| `image` | string | `alpine:latest` | Base image |
| `network` | bool | `false` | Network access |
| `resources` | object | | Resource limits |
| `mounts` | string[] | `[]` | Additional mounts |

#### Triggers

```json
{
  "triggers": [
    {
      "name": "daily-check",
      "type": "cron",
      "schedule": "0 9 * * *",
      "prompt": "Run daily check"
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Trigger name |
| `type` | enum | Yes | `cron`, `watch`, `interval`, `once` |
| `schedule` | string | For cron | Cron expression |
| `path` | string | For watch | Watch path |
| `pattern` | string | For watch | File pattern |
| `interval` | string | For interval | Duration |
| `at` | string | For once | ISO datetime |
| `prompt` | string | Yes | Prompt to send |

#### Delegates

```json
{
  "delegates": ["@helper", "@reviewer"]
}
```

Agents this agent can call via `delegate` tool.

---

## Squad Configuration

`~/.local/share/ayo/sandboxes/squads/{name}/ayo.json`:

```json
{

  "planners": {
    "near_term": "ayo-todos",
    "long_term": "ayo-tickets"
  },
  "agents": {
    "@backend": {
      "model": "your-model",
      "allowed_tools": ["bash", "view", "edit"]
    },
    "@frontend": {
      "model": "your-model"
    }
  },
  "triggers": [],
  "sandbox": {
    "image": "alpine:latest",
    "network": true,
    "resources": {
      "memory": "4G",
      "cpu": "4"
    }
  },
  "memory": {
    "scope": "squad",
    "shared": true
  }
}
```

### Squad Fields

#### Planners

```json
{
  "planners": {
    "near_term": "ayo-todos",
    "long_term": "ayo-tickets"
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `near_term` | string | `ayo-todos` | Session task planner |
| `long_term` | string | `ayo-tickets` | Persistent planner |

#### Agents

Per-agent config overrides:

```json
{
  "agents": {
    "@backend": {
      "model": "your-model",
      "allowed_tools": ["bash", "view", "edit"],
      "memory": {
        "enabled": true
      }
    }
  }
}
```

Overrides merge with agent's base config.

#### Triggers

Squad-level triggers:

```json
{
  "triggers": [
    {
      "name": "standup",
      "type": "cron",
      "schedule": "0 9 * * MON-FRI",
      "agent": "@lead",
      "prompt": "Run standup"
    }
  ]
}
```

#### Memory

```json
{
  "memory": {
    "scope": "squad",
    "shared": true
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | enum | `squad` | Memory scope |
| `shared` | bool | `true` | Share between agents |

---

## JSON Schema

Full JSON Schema available at:
- `schemas/ayo.json` in repository

### Validation

```bash
# Validate agent config
ajv validate -s schemas/agent.json -d ~/.config/ayo/agents/@name/config.json

# Validate squad config
ajv validate -s schemas/squad.json -d squad/ayo.json
```

---

## Type Definitions

### TrustLevel

```typescript
type TrustLevel = "sandboxed" | "privileged" | "unrestricted";
```

### MemoryScope

```typescript
type MemoryScope = "global" | "agent" | "path" | "squad";
```

### MemoryCategory

```typescript
type MemoryCategory = "preference" | "fact" | "correction" | "pattern";
```

### TriggerType

```typescript
type TriggerType = "cron" | "watch" | "interval" | "once" | "daily" | "weekly" | "monthly";
```

### SandboxProvider

```typescript
type SandboxProvider = "applecontainer" | "nspawn";
```

---

## Config Resolution

### Priority (highest to lowest)

1. CLI flags
2. Environment variables
3. Project config (`./.config/ayo/ayo.json`)
4. Agent/Squad config
5. User config (`~/.config/ayo/config.json`)
6. System defaults

### Merging

Configs are deep-merged. Arrays are replaced, not concatenated.

```javascript
// Base config
{ "tools": ["bash", "view"] }

// Override
{ "tools": ["edit"] }

// Result
{ "tools": ["edit"] }  // Replaced, not merged
```
