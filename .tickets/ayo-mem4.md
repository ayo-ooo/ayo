---
id: ayo-mem4
status: open
deps: [ayo-mem1]
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 3
assignee: Alex Cabrera
parent: ayo-memx
tags: [memory, backup]
---
# Add memory export/import

Allow users to backup and restore memories.

## Commands

```bash
# Export all memories to file
ayo memory export memories.json
ayo memory export --scope global global-memories.json

# Import memories from file
ayo memory import memories.json
ayo memory import --merge memories.json  # Don't overwrite existing
ayo memory import --dry-run memories.json  # Show what would be imported
```

## Export Format

```json
{
  "version": "1",
  "exported_at": "2026-02-24T01:30:00Z",
  "memories": [
    {
      "id": "mem_abc123",
      "content": "User prefers TypeScript",
      "category": "preference",
      "scope": "global",
      "agent": null,
      "path": null,
      "created_at": "2026-02-15T10:30:00Z",
      "embedding": null,  // Excluded by default
      "supersedes": "mem_old456",
      "topics": ["typescript", "preferences"]
    }
  ]
}
```

## Options

| Flag | Description |
|------|-------------|
| `--include-embeddings` | Include embedding vectors (larger file) |
| `--scope` | Filter export by scope |
| `--since` | Export memories created after date |
| `--merge` | Merge with existing, don't overwrite |
| `--dry-run` | Show what would be imported |

## Implementation

### Files to Modify

- `cmd/ayo/memory.go` - Add export/import subcommands
- `internal/memory/memory.go` - Add Export/Import methods

## Testing

- Test round-trip export/import
- Test merge behavior
- Test dry-run output
- Test scope filtering
