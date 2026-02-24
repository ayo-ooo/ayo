---
id: ayo-fc4a
status: open
deps: [ayo-spy5, ayo-1v23]
links: []
created: 2026-02-23T22:16:17Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [docs, squads]
---
# Rewrite squads and triggers docs

Rewrite squads documentation and create comprehensive triggers documentation. Include architecture diagrams, examples, and troubleshooting guides.

## Context

Squads and triggers are key features that need clear documentation. This ticket creates/rewrites these docs after the Phase 4 trigger system is complete.

## Files to Create/Rewrite

### docs/squads.md (~500 lines)

Structure:
```markdown
# Squads

## What is a Squad?
[Definition, use cases]

## Architecture
[Diagram of shared sandbox, agent users]

## Creating a Squad
[Step-by-step with examples]

## Squad Configuration (ayo.json)
[Full schema]

## The Constitution (SQUAD.md)
[Purpose, writing guide]

## Squad Lead
[Role, responsibilities, tools]

## Dispatch Routing
[How messages reach agents]

## Ticket Coordination
[How agents coordinate via tickets]

## I/O Schemas
[Input/output validation]

## Workspace Management
[Initialization, permissions]

## Examples
[Common squad patterns]

## Troubleshooting
[Common issues and solutions]
```

### docs/triggers.md (~400 lines)

Structure:
```markdown
# Triggers

## What are Triggers?
[Definition, use cases]

## Trigger Types
### Cron
### Interval
### Daily/Weekly/Monthly
### One-Time
### File Watch

## Configuration
[YAML schema, examples]

## CLI Commands
[create, list, show, enable, disable, delete, test]

## Trigger Engine
[How daemon manages triggers]

## Notifications
[How to get notified of completions]

## Patterns
[Link to docs/patterns/]

## Troubleshooting
[Common issues and solutions]
```

## Architecture Diagrams

Include ASCII diagrams for:

### Squad Architecture
```
┌─────────────────────────────────────────┐
│ Squad Sandbox (shared)                  │
│                                         │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │@architect│ │@frontend│ │@backend │   │
│  │  (lead) │ │         │ │         │   │
│  └────┬────┘ └────┬────┘ └────┬────┘   │
│       │           │           │         │
│       └─────────┬─┴───────────┘         │
│                 │                        │
│           ┌─────┴─────┐                 │
│           │ /workspace │                 │
│           │ /.tickets  │                 │
│           └───────────┘                 │
└─────────────────────────────────────────┘
```

### Trigger Flow
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Trigger     │────▶│ gocron      │────▶│ Agent       │
│ Config      │     │ Scheduler   │     │ Execution   │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                    ┌─────────────┐     ┌──────▼──────┐
                    │ Notification│◀────│ Completion  │
                    │ System      │     │ Handler     │
                    └─────────────┘     └─────────────┘
```

## Acceptance Criteria

- [ ] squads.md is ~500 lines, comprehensive
- [ ] triggers.md is ~400 lines, covers all types
- [ ] Architecture diagrams included
- [ ] Each section has practical examples
- [ ] Troubleshooting covers common issues
- [ ] Links to pattern documentation
- [ ] No references to removed features

## Testing

- Verify all examples work
- Check all links
- Review diagrams for accuracy
