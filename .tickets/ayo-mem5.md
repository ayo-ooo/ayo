---
id: ayo-mem5
status: open
deps: [ayo-mem1, ayo-mem2, ayo-mem3]
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-memx
tags: [memory, docs]
---
# Document memory system

Comprehensive documentation for the memory system.

## Documentation Structure

### docs/memory.md

```markdown
# Memory System

Ayo's memory system allows agents to learn and adapt over time...

## How It Works

### Automatic Formation
Agents automatically extract memorable content from conversations...

### Memory Categories
- Preference: User preferences (tools, styles, workflows)
- Fact: Facts about user, project, or environment
- Correction: User corrections to agent behavior
- Pattern: Observed behavioral patterns

### Memory Scopes
- Global: Applies to all agents, all directories
- Agent: Specific to one agent
- Path: Specific to a project directory
- Squad: Shared across squad agents

## Managing Memories

### CLI Commands
...

### Agent Tools
...

## Configuration

### Per-Agent Memory Settings
...

### Squad Memory Sharing
...

## Technical Details

### Storage
- SQLite for fast semantic search
- Zettelkasten markdown for human readability

### Deduplication
...

### Supersession
...
```

## Also Update

- `docs/TUTORIAL.md` - Add memory section
- `docs/agents.md` - Reference memory config
- `docs/squads.md` - Document squad memory sharing
- `AGENTS.md` - Update memory references

## Testing

- Review docs for clarity
- Test all documented commands work
- Verify config examples are accurate
