# Memory System

Ayo's memory system allows agents to learn and adapt over time by remembering facts, preferences, patterns, and corrections from conversations.

## How It Works

### Automatic Formation

Agents automatically extract memorable content from conversations using the small model. When an agent identifies something worth remembering (a user preference, a project fact, or a pattern), it stores it as a memory for future reference.

### Memory Categories

| Category | Description | Example |
|----------|-------------|---------|
| `preference` | User preferences | "User prefers TypeScript over JavaScript" |
| `fact` | Facts about user/project | "Project uses PostgreSQL database" |
| `correction` | User corrections | "Don't suggest deprecated APIs" |
| `pattern` | Observed patterns | "User tends to work on tests in the afternoon" |

### Memory Scopes

| Scope | Description |
|-------|-------------|
| Global | Applies to all agents, all directories |
| Agent | Specific to one agent (e.g., only @ayo sees it) |
| Path | Specific to a project directory |
| Squad | Shared across all agents in a squad |

## Managing Memories

### CLI Commands

```bash
# List all memories
ayo memory list

# Search memories semantically
ayo memory search "user preferences"

# Show a specific memory
ayo memory show <id>

# Store a new memory manually
ayo memory store "User prefers dark themes" --category preference

# Forget (soft delete) a memory
ayo memory forget <id>

# Get memory statistics
ayo memory stats

# Clear all memories (with confirmation)
ayo memory clear

# Export memories for backup
ayo memory export backup.json
ayo memory export --agent @ayo agent-memories.json
ayo memory export --include-embeddings full-backup.json

# Import memories from backup
ayo memory import backup.json
ayo memory import --merge backup.json      # Don't overwrite existing
ayo memory import --dry-run backup.json    # Preview what would be imported
```

### Agent Memory Tools

Agents have tools to manage memories during conversations:

```json
// Store a memory
{
  "name": "memory_store",
  "arguments": {
    "content": "User prefers vim keybindings",
    "category": "preference"
  }
}

// Search memories
{
  "name": "memory_search",
  "arguments": {
    "query": "user preferences for editor"
  }
}

// Recall memories (loaded into context)
{
  "name": "memory_recall",
  "arguments": {
    "query": "current project setup"
  }
}
```

## Configuration

### Per-Agent Memory Settings

In an agent's `ayo.json` or `config.json`:

```json
{
  "agent": {
    "memory": {
      "enabled": true,
      "auto_formation": true,
      "embedding_model": "nomic-embed-text",
      "search_threshold": 0.7
    }
  }
}
```

### Global Memory Settings

In `~/.config/ayo/config.toml`:

```toml
[memory]
enabled = true
auto_formation = true
embedding_provider = "ollama"
embedding_model = "nomic-embed-text"
```

### Squad Memory Sharing

Squads share memories by default. All agents in a squad can access squad-scoped memories:

```yaml
# In SQUAD.md frontmatter
---
memory:
  shared: true           # Agents share squad memories
  inherit_global: true   # Squad agents also see global memories
---
```

## Technical Details

### Storage

Ayo uses dual storage for memories:

1. **SQLite** - Primary storage with vector search for semantic retrieval
2. **Zettelkasten** - Optional markdown files for human readability

### Semantic Search

Memories use embedding vectors for semantic search. When you search for "user preferences", it finds memories that are semantically similar, not just keyword matches.

Default embedding model: `nomic-embed-text` (via Ollama)

### Deduplication

When storing new memories, the system checks for duplicates using:
1. Semantic similarity (>95% similar = potential duplicate)
2. Content hash (exact duplicates)

### Supersession

Newer memories can supersede older ones. When information is updated:
1. The new memory is created
2. The old memory is marked as `superseded`
3. The chain is maintained for history

```bash
# View supersession chain
ayo memory show <id> --history
```

## Export Format

```json
{
  "version": "1",
  "exported_at": "2026-02-24T12:00:00Z",
  "memories": [
    {
      "id": "mem_abc123",
      "content": "User prefers TypeScript",
      "category": "preference",
      "agent": null,
      "path": null,
      "created_at": "2026-02-15T10:30:00Z",
      "confidence": 0.95,
      "supersedes": null,
      "embedding": null
    }
  ]
}
```

## Privacy & Security

- Memories are stored locally in `~/.local/share/ayo/ayo.db`
- Embeddings are generated locally (via Ollama)
- No memory data is sent to cloud services
- Export/import allows you to control your data
