---
id: ayo-spy5
status: open
deps: [ayo-6f6b]
links: []
created: 2026-02-23T22:16:17Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [docs, core]
---
# Rewrite core documentation

Rewrite the core documentation files to reflect the new simplified architecture. Focus on clarity, brevity, and practical examples.

## Context

After Phase 1-5 changes, the documentation is outdated. This ticket rewrites the core docs to match the new architecture.

## Files to Rewrite

### README.md (~200 lines)

Structure:
```markdown
# ayo

> AI agents that live on your machine

## Quick Start
[5-line install + first run]

## Features
- Agents in isolated sandboxes
- Squads for team coordination
- Triggers for ambient agents
- Host file access with approval

## Documentation
[Links to docs/]

## Examples
[3-4 common use cases]

## License
```

### docs/getting-started.md (~300 lines)

Structure:
```markdown
# Getting Started

## Installation
[Platform-specific instructions]

## First Agent
[Create and run @crush]

## Understanding Sandboxes
[Brief explanation of isolation]

## Your First Squad
[Create a simple squad]

## Next Steps
[Links to deeper docs]
```

### docs/agents.md (~400 lines)

Structure:
```markdown
# Agents

## What is an Agent?
[Definition, components]

## Agent Configuration (ayo.json)
[Full schema with examples]

## Agent Files
[AGENT.md, skills/, etc.]

## Built-in Agents
[@ayo, @crush descriptions]

## Creating Custom Agents
[Step-by-step guide]

## Tools
[Overview of available tools]

## Permissions
[file_request, auto_approve]
```

## Guidelines

- Remove all references to removed features (flows, plugins complexity)
- Use concrete examples throughout
- Keep language direct and scannable
- Include common troubleshooting tips
- Link related docs appropriately

## Acceptance Criteria

- [ ] README.md is ~200 lines, focused on value proposition
- [ ] getting-started.md is ~300 lines, gets user running in 5 minutes
- [ ] agents.md is ~400 lines, comprehensive but scannable
- [ ] No references to removed features
- [ ] All code examples are tested and work
- [ ] Links between docs are valid
- [ ] Consistent formatting throughout

## Testing

- Run through getting-started as new user
- Verify all code examples work
- Check all links are valid
- Have someone unfamiliar review for clarity
