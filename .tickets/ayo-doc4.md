---
id: ayo-doc4
status: open
deps: [ayo-doc2, ayo-doc3]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-docs
tags: [documentation, phase9]
---
# Task: Write Tutorials

## Summary

Write 5 hands-on tutorials in `docs/tutorials/` that guide users through common workflows.

## Tutorials

### 1. first-agent.md
**Goal**: Create and customize your first agent

```markdown
# Create Your First Agent

## What You'll Build
A custom agent specialized for code review.

## Prerequisites
- ayo installed and setup complete

## Steps
1. Create agent directory
2. Write system.md
3. Configure ayo.json
4. Add tools
5. Test the agent

## Complete Example
[Full working example]

## Next Steps
[Links to squads, triggers]
```

### 2. squads.md
**Goal**: Set up a multi-agent squad

```markdown
# Multi-Agent Coordination with Squads

## What You'll Build
A development squad with frontend and backend agents.

## Prerequisites
- first-agent tutorial complete

## Steps
1. Create squad directory
2. Write SQUAD.md
3. Configure agents
4. Run a coordinated task
5. Monitor progress

## Complete Example
[Full working example]
```

### 3. triggers.md
**Goal**: Set up event-driven agents

```markdown
# Event-Driven Agents with Triggers

## What You'll Build
An agent that runs on a schedule.

## Prerequisites
- first-agent tutorial complete

## Steps
1. Configure cron trigger
2. Test trigger execution
3. Add file watch trigger
4. Monitor trigger history

## Complete Example
[Full working example]
```

### 4. memory.md
**Goal**: Use the memory system

```markdown
# Memory System Deep Dive

## What You'll Build
An agent that remembers your preferences.

## Prerequisites
- first-agent tutorial complete

## Steps
1. Enable memory for an agent
2. Create memories
3. Query memories
4. Use memory scopes
5. Export/backup memories

## Complete Example
[Full working example]
```

### 5. plugins.md
**Goal**: Create a custom plugin

```markdown
# Creating Plugins

## What You'll Build
A plugin with a custom tool.

## Prerequisites
- first-agent tutorial complete

## Steps
1. Create plugin directory
2. Write manifest.json
3. Create a tool
4. Install plugin
5. Use in agents

## Complete Example
[Full working example]
```

## Requirements

- Each tutorial completable in < 30 minutes
- All code examples tested and working
- Clear before/after states
- Troubleshooting section in each
- Cross-references to relevant guides

## Success Criteria

- [ ] All 5 tutorials complete
- [ ] All examples tested
- [ ] Consistent structure across tutorials
- [ ] Progressive complexity

---

*Created: 2026-02-23*
