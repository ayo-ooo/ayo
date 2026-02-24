---
id: ayo-nqyv
status: closed
deps: []
links: []
created: 2026-02-23T23:13:19Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-pv3a
tags: [schema, squads]
---
# Design ayo.json schema with squad namespace

Extend ayo.json schema with `squad` namespace for squad configuration.

## Schema Design

```json
{
  "$schema": "https://ayo.dev/schemas/ayo.json",
  "version": "1",
  
  "squad": {
    "description": "Development team for auth features",
    "lead": "@architect",
    "input_accepts": "@planner",
    "agents": ["@frontend", "@backend", "@qa"],
    "planners": {
      "near_term": "ayo-todos",
      "long_term": "ayo-tickets"
    },
    "sandbox": {
      "image": "alpine:3.21",
      "network": true,
      "mounts": ["~/Projects/myapp:/workspace"]
    },
    "io": {
      "input_schema": "schemas/input.json",
      "output_schema": "schemas/output.json"
    },
    "coordination": {
      "ticket_workflow": "kanban",
      "auto_assign": true
    }
  }
}
```

## Field Definitions

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Human-readable squad description |
| `lead` | string | Agent that receives all messages and routes |
| `input_accepts` | string | Agent that processes squad input (if different from lead) |
| `agents` | string[] | Agents available in this squad |
| `planners.near_term` | string | Short-term planner (e.g., "ayo-todos") |
| `planners.long_term` | string | Long-term planner (e.g., "ayo-tickets") |
| `sandbox.image` | string | Container base image |
| `sandbox.network` | bool | Allow network access |
| `sandbox.mounts` | string[] | Host paths to mount |
| `io.input_schema` | string | JSON Schema for squad input validation |
| `io.output_schema` | string | JSON Schema for squad output validation |
| `coordination.ticket_workflow` | string | "kanban", "scrum", or custom |
| `coordination.auto_assign` | bool | Auto-assign tickets to agents |

## Go Struct Updates

Update `internal/config/config.go`:

```go
type SquadConfig struct {
    Description   string            `json:"description,omitempty"`
    Lead          string            `json:"lead,omitempty"`
    InputAccepts  string            `json:"input_accepts,omitempty"`
    Agents        []string          `json:"agents,omitempty"`
    Planners      *PlannersConfig   `json:"planners,omitempty"`
    Sandbox       *SandboxConfig    `json:"sandbox,omitempty"`
    IO            *IOConfig         `json:"io,omitempty"`
    Coordination  *CoordinationConfig `json:"coordination,omitempty"`
}

type PlannersConfig struct {
    NearTerm string `json:"near_term,omitempty"`
    LongTerm string `json:"long_term,omitempty"`
}

type IOConfig struct {
    InputSchema  string `json:"input_schema,omitempty"`
    OutputSchema string `json:"output_schema,omitempty"`
}

type CoordinationConfig struct {
    TicketWorkflow string `json:"ticket_workflow,omitempty"`
    AutoAssign     bool   `json:"auto_assign,omitempty"`
}
```

## SQUAD.md Changes

With ayo.json handling config, SQUAD.md becomes pure documentation:

```markdown
# Squad: auth-team

## Mission
Implement secure authentication for the e-commerce platform.

## Agents

### @architect (Lead)
Decomposes tasks, reviews output, makes architectural decisions.

### @backend
Implements API endpoints, writes tests.

## Coordination Rules
- All work flows through tickets
- @architect reviews all changes before merge
```

**No more YAML frontmatter in SQUAD.md!**

## Files to Create/Modify

1. Update `schemas/ayo.json` with squad namespace
2. Update `internal/config/config.go` - SquadConfig struct
3. Update `internal/squads/context.go` - ConstitutionFrontmatter removal

## Testing

- Validate squad configs with various field combinations
- Test required vs optional fields
- Test schema validation errors
