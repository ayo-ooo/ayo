---
id: ayo-zett
status: open
deps: [ayo-mem1]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-memx
tags: [memory, zettelkasten]
---
# Task: Zettelkasten Memory Tools & Embedding Links

## Summary

Expose the Zettelkasten memory system through proper tools so agents can manage notes without direct file manipulation. Add metadata linking between SQLite embeddings and Zettelkasten markdown files.

## Context

The memory system stores both:
1. **SQLite** - Embeddings for semantic search
2. **Zettelkasten** - Human-readable markdown files

Currently these are loosely coupled. We need:
- Tools for agents to create/read/link notes
- Metadata in SQLite linking embeddings to note files
- Bidirectional linking between notes (Zettelkasten-style)

## Tool Specifications

### memory_note_create

```json
{
  "name": "memory_note_create",
  "description": "Create a new Zettelkasten memory note",
  "parameters": {
    "title": { "type": "string", "required": true },
    "content": { "type": "string", "required": true },
    "category": { 
      "type": "string", 
      "enum": ["preference", "fact", "correction", "pattern"],
      "required": true 
    },
    "tags": { "type": "array", "items": { "type": "string" } },
    "links": { 
      "type": "array", 
      "items": { "type": "string" },
      "description": "IDs of related notes to link to"
    },
    "scope": {
      "type": "string",
      "enum": ["global", "agent", "path"],
      "default": "global"
    }
  }
}
```

### memory_note_link

```json
{
  "name": "memory_note_link",
  "description": "Link two memory notes together",
  "parameters": {
    "from_id": { "type": "string", "required": true },
    "to_id": { "type": "string", "required": true },
    "relationship": { 
      "type": "string",
      "description": "Type of relationship (e.g., 'supersedes', 'relates_to', 'contradicts')"
    }
  }
}
```

### memory_note_search

```json
{
  "name": "memory_note_search",
  "description": "Search memory notes semantically or by metadata",
  "parameters": {
    "query": { "type": "string", "description": "Semantic search query" },
    "tags": { "type": "array", "items": { "type": "string" } },
    "category": { "type": "string" },
    "scope": { "type": "string" },
    "include_linked": { "type": "boolean", "default": true },
    "limit": { "type": "integer", "default": 10 }
  }
}
```

### memory_note_read

```json
{
  "name": "memory_note_read",
  "description": "Read a specific memory note by ID",
  "parameters": {
    "id": { "type": "string", "required": true },
    "include_links": { "type": "boolean", "default": true }
  }
}
```

### memory_note_update

```json
{
  "name": "memory_note_update",
  "description": "Update an existing memory note",
  "parameters": {
    "id": { "type": "string", "required": true },
    "content": { "type": "string" },
    "tags": { "type": "array", "items": { "type": "string" } },
    "add_links": { "type": "array", "items": { "type": "string" } },
    "remove_links": { "type": "array", "items": { "type": "string" } }
  }
}
```

## SQLite Schema Extensions

```sql
-- Add metadata linking embeddings to notes
ALTER TABLE embeddings ADD COLUMN note_id TEXT;
ALTER TABLE embeddings ADD COLUMN note_path TEXT;

-- Note links table
CREATE TABLE note_links (
    id TEXT PRIMARY KEY,
    from_note_id TEXT NOT NULL,
    to_note_id TEXT NOT NULL,
    relationship TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_note_id, to_note_id)
);

-- Index for fast link lookups
CREATE INDEX idx_note_links_from ON note_links(from_note_id);
CREATE INDEX idx_note_links_to ON note_links(to_note_id);
```

## Zettelkasten File Format

```markdown
---
id: mem-abc123
title: User prefers Go over Python
category: preference
scope: global
tags: [languages, coding-style]
links:
  - mem-def456  # relates_to
  - mem-ghi789  # supersedes
created: 2026-02-23T10:30:00Z
updated: 2026-02-23T10:30:00Z
---

# User prefers Go over Python

When given a choice of implementation language, prefer Go.

## Context

User mentioned this explicitly on 2026-02-20 when asked about
a new CLI tool implementation.

## Links

- [[mem-def456]] - User's Go coding conventions
- [[mem-ghi789]] - Supersedes earlier Python preference
```

## Implementation Steps

1. [ ] Design note ID format (e.g., `mem-{nanoid}`)
2. [ ] Add `note_id` and `note_path` columns to embeddings table
3. [ ] Create `note_links` table
4. [ ] Implement `memory_note_create` tool
5. [ ] Implement `memory_note_link` tool
6. [ ] Implement `memory_note_search` tool
7. [ ] Implement `memory_note_read` tool
8. [ ] Implement `memory_note_update` tool
9. [ ] Update embedding creation to link to notes
10. [ ] Add backlink resolution in search results
11. [ ] Update CLI `ayo memory` commands to use same tools
12. [ ] Write tests

## Dependencies

- Depends on: `ayo-mem1` (memory CLI commands)
- Blocks: `ayo-mem3` (squad-scoped memories)

## Acceptance Criteria

- [ ] Agents can create notes without file manipulation
- [ ] Notes link to embeddings via SQLite metadata
- [ ] Bidirectional links work (from → to and to → from)
- [ ] Search returns linked notes when requested
- [ ] CLI and agent tools use same underlying implementation
- [ ] Human-readable Zettelkasten files remain browsable

## Notes

- Keep file paths predictable: `~/.local/share/ayo/memory/notes/{scope}/{category}/{id}.md`
- Consider graph visualization of note links (future)
- Links should survive note moves/renames via ID-based references

---

*Created: 2026-02-23*
