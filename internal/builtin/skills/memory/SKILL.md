---
name: memory
description: Guidelines for using the memory tool to store and retrieve persistent information across sessions.
metadata:
  author: ayo
  version: "1.1"
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

## Memory Lifecycle

- Memories persist across sessions
- Similar memories are automatically deduplicated or superseded
- Users can manage memories via `ayo memory` CLI commands
- Outdated memories can be forgotten via the forget operation
