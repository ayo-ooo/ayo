# Squad Configuration Guide

Complete reference for configuring ayo squads.

## Directory Structure

```
~/.local/share/ayo/sandboxes/squads/{name}/
├── SQUAD.md              # Required: Team constitution
├── ayo.json              # Optional: Squad configuration
├── workspace/            # Shared code workspace
├── .tickets/             # Coordination tickets
├── .context/             # Session persistence
│   └── session.json
└── schemas/              # Optional: I/O schemas
    ├── input.json
    └── output.json
```

## SQUAD.md Format

The constitution defines the squad's mission and coordination rules.

### Basic Structure

```markdown
---
name: dev-team
planners:
  near_term: ayo-todos
  long_term: ayo-tickets
agents:
  - "@backend"
  - "@frontend"
lead: "@backend"
---

# Squad: dev-team

## Mission

Brief description of squad's purpose.

## Agents

### @backend
- Responsibilities
- Expertise areas

### @frontend
- Responsibilities
- Expertise areas

## Workflow

How agents coordinate.

## Coordination Rules

- Rules for ticket handoffs
- Communication patterns
```

### Frontmatter Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Squad identifier |
| `planners` | object | No | Near-term and long-term planners |
| `agents` | string[] | No | Agent handles (or extract from markdown) |
| `lead` | string | No | Default dispatch target |
| `input_accepts` | object | No | Input routing patterns |

### Planners

```yaml
planners:
  near_term: ayo-todos    # Session task tracking
  long_term: ayo-tickets  # Persistent coordination
```

**Available planners**:
- `ayo-todos` - In-memory task list (session-scoped)
- `ayo-tickets` - File-based tickets (persistent)

### Input Routing

```yaml
input_accepts:
  "@backend":
    - "*api*"
    - "*database*"
    - "*server*"
  "@frontend":
    - "*ui*"
    - "*component*"
    - "*react*"
```

When input matches a pattern, route to that agent.

## ayo.json Schema

```json
{
  "planners": {
    "near_term": "ayo-todos",
    "long_term": "ayo-tickets"
  },
  "agents": {
    "@backend": {
      "model": "claude-sonnet-4-20250514",
      "allowed_tools": ["bash", "view", "edit"]
    },
    "@frontend": {
      "model": "claude-sonnet-4-20250514",
      "allowed_tools": ["bash", "view", "edit"]
    }
  },
  "triggers": [
    {
      "name": "daily-standup",
      "type": "cron",
      "schedule": "0 9 * * MON-FRI",
      "agent": "@lead",
      "prompt": "Run standup"
    }
  ],
  "sandbox": {
    "image": "alpine:latest",
    "network": true,
    "resources": {
      "memory": "4G"
    }
  },
  "memory": {
    "scope": "squad",
    "shared": true
  }
}
```

### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `planners` | object | Planner configuration |
| `agents` | object | Per-agent config overrides |
| `triggers` | array | Squad-level triggers |
| `sandbox` | object | Sandbox configuration |
| `memory` | object | Memory sharing settings |

## Agent Configuration in Squads

### Override Agent Settings

In `ayo.json`:

```json
{
  "agents": {
    "@backend": {
      "model": "claude-sonnet-4-20250514",
      "allowed_tools": ["bash", "view", "edit", "grep"],
      "memory": {
        "enabled": true,
        "scope": "squad"
      }
    }
  }
}
```

Squad config merges with (and overrides) agent's base config.

### Agent Users

Each agent gets a Unix user in the sandbox:
- Username: Sanitized handle (e.g., `backend` for `@backend`)
- Home: `/home/{username}/`
- Shared workspace: `/workspaces/{squad}/`

## I/O Schemas

### Input Schema

`schemas/input.json`:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "task": {
      "type": "string",
      "description": "Task to perform"
    },
    "priority": {
      "type": "string",
      "enum": ["low", "medium", "high"]
    }
  },
  "required": ["task"]
}
```

### Output Schema

`schemas/output.json`:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["success", "failed", "partial"]
    },
    "summary": {
      "type": "string"
    },
    "artifacts": {
      "type": "array",
      "items": {
        "type": "string"
      }
    }
  }
}
```

## Ticket System

### Ticket Format

`.tickets/feat-001.md`:

```markdown
---
id: feat-001
status: open
assignee: "@backend"
deps: []
priority: high
created: 2024-01-15T10:00:00Z
---
# Implement user API

## Description

Create CRUD endpoints for users.

## Acceptance Criteria

- [ ] GET /api/users
- [ ] POST /api/users
- [ ] PUT /api/users/:id
- [ ] DELETE /api/users/:id
```

### Ticket Status

| Status | Description |
|--------|-------------|
| `open` | Not started |
| `in_progress` | Being worked on |
| `blocked` | Waiting on dependencies |
| `review` | Ready for review |
| `closed` | Completed |

### Dependencies

```yaml
deps:
  - "feat-000"  # Must be closed first
  - "setup-db"
```

Ticket is `blocked` until all deps are closed.

## Dispatch Routing

When you send `ayo "#squad" "task"`:

1. **Explicit target**: `ayo "#squad" @agent "task"` → routes to @agent
2. **Input matching**: Check `input_accepts` patterns
3. **Lead agent**: Route to `lead` if defined
4. **Default**: Route to `@ayo` for orchestration

### Routing Configuration

In SQUAD.md frontmatter:

```yaml
lead: "@backend"
input_accepts:
  "@backend":
    - "*api*"
    - "*database*"
  "@frontend":
    - "*ui*"
    - "*component*"
```

## Memory Sharing

### Squad-Scoped Memory

All agents share squad memories:

```json
{
  "memory": {
    "scope": "squad",
    "shared": true
  }
}
```

### Memory Isolation

Each agent has private memories:

```json
{
  "memory": {
    "scope": "agent",
    "shared": false
  }
}
```

## Session Context

`.context/session.json`:

```json
{
  "last_session_id": "sess_abc123",
  "session_count": 42,
  "agent_memories": {
    "@backend": {
      "last_task": "Implemented auth",
      "working_on": "Database migration"
    }
  },
  "notes": [
    "Decided to use PostgreSQL",
    "API versioning with /v2/"
  ]
}
```

Context persists across sessions for continuity.

## Complete Example

### Full Squad Configuration

**SQUAD.md**:
```markdown
---
name: fullstack-team
planners:
  near_term: ayo-todos
  long_term: ayo-tickets
lead: "@architect"
agents:
  - "@architect"
  - "@backend"
  - "@frontend"
  - "@qa"
input_accepts:
  "@backend":
    - "*api*"
    - "*database*"
    - "*server*"
  "@frontend":
    - "*ui*"
    - "*component*"
    - "*css*"
  "@qa":
    - "*test*"
    - "*bug*"
---

# Squad: fullstack-team

## Mission

Build production-ready features through coordinated full-stack development.

## Agents

### @architect
Lead architect responsible for:
- System design decisions
- Code review for architecture
- Breaking work into tickets
- Coordinating agent handoffs

### @backend
Backend developer handling:
- API design and implementation
- Database operations
- Authentication/authorization
- Performance optimization

### @frontend
Frontend developer handling:
- React component development
- State management
- API integration
- Responsive design

### @qa
Quality engineer handling:
- Test planning
- E2E test implementation
- Bug tracking
- Performance testing

## Workflow

1. @architect receives task, creates tickets
2. @backend implements APIs (creates "api-ready" tickets)
3. @frontend builds UI after APIs ready
4. @qa tests integration
5. @architect reviews and approves

## Coordination Rules

- All work tracked in tickets
- Use dependency links for sequencing
- API specs documented before UI work
- No feature complete without tests
```

**ayo.json**:
```json
{
  "planners": {
    "near_term": "ayo-todos",
    "long_term": "ayo-tickets"
  },
  "sandbox": {
    "image": "alpine:latest",
    "network": true,
    "resources": {
      "memory": "8G",
      "cpu": "4"
    }
  },
  "memory": {
    "scope": "squad",
    "shared": true
  }
}
```

## Troubleshooting

### Squad not starting

```bash
# Check daemon status
ayo sandbox service status

# Check sandbox provider
ayo doctor | grep sandbox

# View logs
tail -f ~/.local/share/ayo/daemon.log
```

### Agent not receiving tasks

1. Verify agent is in SQUAD.md `agents` list
2. Check `input_accepts` patterns if used
3. Try explicit routing: `ayo "#squad" @agent "task"`

### Tickets not syncing

Tickets are inside sandbox at `/.tickets/`:

```bash
ayo squad shell my-squad
ls /.tickets/
```

### Memory not shared

Ensure `memory.scope` is `squad` and `shared` is `true`.
