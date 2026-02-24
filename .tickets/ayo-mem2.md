---
id: ayo-mem2
status: closed
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-memx
tags: [memory, tools]
---
# Add memory tools for agents

Give agents explicit tools to store and search memories, beyond automatic formation.

## Tools

### memory_store

```json
{
  "name": "memory_store",
  "description": "Store an important piece of information for future reference",
  "parameters": {
    "content": {
      "type": "string",
      "description": "The information to remember"
    },
    "category": {
      "type": "string",
      "enum": ["preference", "fact", "correction", "pattern"],
      "description": "Category of the memory"
    },
    "scope": {
      "type": "string",
      "enum": ["global", "agent", "path"],
      "description": "Scope of the memory (default: hybrid)"
    }
  },
  "required": ["content"]
}
```

### memory_search

```json
{
  "name": "memory_search",
  "description": "Search for relevant memories",
  "parameters": {
    "query": {
      "type": "string",
      "description": "Search query"
    },
    "limit": {
      "type": "integer",
      "description": "Maximum results (default: 5)"
    },
    "scope": {
      "type": "string",
      "enum": ["global", "agent", "path", "all"],
      "description": "Scope to search (default: all)"
    }
  },
  "required": ["query"]
}
```

## Use Cases

1. **Agent learns user preference**: "I noticed you always use tabs. Let me remember that."
2. **Agent stores project fact**: "This project uses custom ESLint rules, storing for reference."
3. **Agent recalls context**: Before starting work, agent searches for relevant memories.

## Implementation

### Files to Create

- `internal/tools/memory_store.go`
- `internal/tools/memory_search.go`

### Integration

Register tools in `internal/tools/registry.go`.

## Testing

- Test memory creation via tool
- Test duplicate detection still works
- Test search returns relevant results
- Test scope enforcement
