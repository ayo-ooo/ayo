---
name: memory
description: Guidelines for using the memory tool to store and retrieve persistent information across sessions.
metadata:
  author: ayo
  version: "2.0"
---

# Memory Management

You have access to a memory tool for storing and retrieving persistent information.

## When to Store Memories

Store information when users:
- Express preferences ("I prefer X", "always use Y", "never do Z")
- Correct your behavior ("no, I meant...", "actually...", "that's wrong")
- Share facts about themselves or their project
- Explicitly request you remember something

## Memory Categories

Categories are auto-detected if not specified:

- **preference**: User preferences about tools, styles, workflows
- **fact**: Facts about the user, project, or environment
- **correction**: Corrections to agent behavior
- **pattern**: Observed behavioral patterns

## Best Practices

### Storing Memories
1. **Search first** to avoid duplicates
2. **Distill to essence**: "User prefers TypeScript" not "The user mentioned they prefer TypeScript over JavaScript"
3. **Be specific**: Include enough context to be useful
4. **Let auto-categorization work**: Only specify category if you disagree with the default

### Searching Memories
1. Search at the start of tasks that might have relevant context
2. Use semantic queries, not exact matches
3. Search when users ask about their preferences

### Examples

**User says**: "I always want you to use pnpm instead of npm"
**Action**: 
```json
{"operation": "store", "content": "User prefers pnpm over npm for package management"}
```

**User asks**: "What's my preferred package manager?"
**Action**:
```json
{"operation": "search", "query": "package manager preference"}
```

**User says**: "Remember that this project uses PostgreSQL 15"
**Action**:
```json
{"operation": "store", "content": "Project uses PostgreSQL 15"}
```

## Memory Storage

Memories are stored as Markdown files with TOML frontmatter:

```markdown
+++
id = "mem_01HX..."
created = 2024-01-15T10:30:00Z
category = "preference"
topics = ["go", "testing"]
source = "session:abc123"
+++

User strongly prefers table-driven tests in Go.
```

### Directory Structure

```
~/.local/share/ayo/memory/
├── index.md              # Overview (auto-generated)
├── preferences/          # User preferences
├── facts/                # Facts about user/project
├── corrections/          # Behavior corrections
├── patterns/             # Observed patterns
└── .index.sqlite         # Search index (derived)
```

## Memory Scopes

Memories can have different scopes:

| Scope | Applies To |
|-------|-----------|
| Global | All agents, all projects |
| Agent | Specific agent only |
| Path | Specific directory/project |

## Automatic Memory Features

### Auto-Injection
Relevant memories are automatically injected into system prompts at session start based on semantic similarity.

### Auto-Detection
The system can automatically detect memory-worthy content during conversations using a small local model.

### Conflict Resolution
When new memories conflict with existing ones:
- More recent takes precedence
- Conflicting memories are marked for review
- Agent may ask for clarification if unclear

## Memory Lifecycle

- Memories persist as Markdown files
- Similar memories are automatically deduplicated or superseded
- Users can manage memories via `ayo memory` CLI commands
- Outdated memories can be forgotten via the forget operation
- Index is rebuilt automatically when files change

## CLI Commands

```bash
ayo memory list              # List all memories
ayo memory search "query"    # Semantic search
ayo memory show <id>         # View memory details
ayo memory store "content"   # Store new memory
ayo memory forget <id>       # Remove a memory
ayo memory reindex           # Rebuild search index
```
