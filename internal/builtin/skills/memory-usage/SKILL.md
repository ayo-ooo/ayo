---
name: memory-usage
description: Understanding and using the memory system effectively. Use when storing user preferences, recalling context, or managing long-term knowledge.
compatibility: Requires memory tool and configured embedding provider
metadata:
  author: ayo
  version: "1.0"
---

# Memory Usage Skill

Use the memory system to store, retrieve, and manage persistent knowledge across sessions.

## When to Use

Activate this skill when:
- User shares preferences or corrections
- Important facts should be remembered for future sessions
- User asks "remember this" or "don't forget"
- Context from past interactions would improve current response
- Managing stored memories (listing, forgetting)

## Memory Categories

Memories are automatically categorized:

| Category | When to Store | Examples |
|----------|---------------|----------|
| `preference` | User expresses how they like things | "I prefer tabs over spaces", "Use British English" |
| `fact` | Objective information about user/project | "I work at Acme Corp", "This project uses Go 1.22" |
| `correction` | User corrects agent behavior | "No, I said X not Y", "That's wrong because..." |
| `pattern` | Observed behavioral patterns | User always runs tests before commits |

## Using the Memory Tool

### Storing Memories

Store when information should persist across sessions:

```json
{
  "action": "store",
  "content": "User prefers concise responses without code comments"
}
```

Category is auto-detected but can be explicit:
```json
{
  "action": "store",
  "content": "Project uses pnpm instead of npm",
  "category": "fact"
}
```

### Searching Memories

Find relevant context for current task:

```json
{
  "action": "search",
  "query": "testing preferences"
}
```

### Listing Memories

View all stored memories:

```json
{
  "action": "list"
}
```

### Forgetting Memories

Remove outdated or incorrect memories:

```json
{
  "action": "forget",
  "id": "mem_abc123"
}
```

## Memory Scopes

Memories can have different scopes:

| Scope | Applies To | Use For |
|-------|-----------|---------|
| Global | All agents, all projects | Universal preferences |
| Agent | Specific agent only | Agent-specific behavior |
| Path | Specific directory | Project-specific facts |

## What to Remember

### Good Candidates for Memory

- User preferences (coding style, communication style)
- Project-specific facts (tech stack, conventions)
- Corrections to agent behavior
- Important context that improves future interactions

### Poor Candidates for Memory

- Temporary information (session-specific context)
- Obvious defaults (widely-known best practices)
- Large data (code snippets, full files)
- Sensitive information (credentials, secrets)

## Best Practices

1. **Be concise** - Store the essence, not full conversations
2. **Be specific** - "Uses Go 1.22" not "Uses Go"
3. **Avoid duplicates** - Search before storing similar info
4. **Clean up** - Forget outdated memories when noticed
5. **Scope appropriately** - Don't make project-specific facts global

## Automatic Memory Formation

Some agents auto-detect memorable content during conversation:
- Explicit corrections ("No, I meant...")
- Stated preferences ("I prefer...", "I always...")
- Project facts ("This project uses...", "Our team...")

## Memory Retrieval

Relevant memories are automatically injected into system prompts:
- Semantic search finds related memories
- Top matches included as context
- Threshold and limits configurable per agent

## Troubleshooting

### Memory Not Found

```json
{
  "action": "search",
  "query": "broader search terms"
}
```

### Too Many Irrelevant Memories

Clean up with targeted forget:
```json
{
  "action": "list"
}
```
Then forget outdated entries.

### Memory Not Being Used

Check agent configuration:
- `memory.enabled: true`
- `memory.retrieval.auto_inject: true`
- Memory tool in `allowed_tools`
