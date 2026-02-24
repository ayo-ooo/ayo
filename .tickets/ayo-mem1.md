---
id: ayo-mem1
status: closed
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-memx
tags: [memory, cli]
---
# Add memory CLI commands

Expose the memory system through CLI commands for user visibility and management.

## Commands

```bash
# List memories (with filters)
ayo memory list
ayo memory list --scope global
ayo memory list --scope agent --agent @crush
ayo memory list --scope path --path ~/Projects/myapp
ayo memory list --category preference
ayo memory list --search "typescript"

# View a specific memory
ayo memory show <id>

# Add a memory manually
ayo memory add "User prefers Go for backend services"
ayo memory add --category fact "Project uses PostgreSQL 15"
ayo memory add --scope agent --agent @crush "Always run tests after changes"

# Delete a memory
ayo memory delete <id>

# Search memories semantically
ayo memory search "database configuration"
```

## Implementation

### Files to Create

- `cmd/ayo/memory.go` - CLI command handlers

### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List memories with optional filters |
| `show` | Show single memory with full details |
| `add` | Add a new memory |
| `delete` | Delete a memory |
| `search` | Semantic search across memories |

### Output Format

```
$ ayo memory list --limit 5
ID          CATEGORY     SCOPE    CONTENT
mem_abc123  preference   global   User prefers TypeScript for frontend
mem_def456  fact         path     Project uses PostgreSQL 15
mem_ghi789  correction   agent    @crush should use npm, not yarn
...

$ ayo memory show mem_abc123
ID:         mem_abc123
Category:   preference
Scope:      global
Created:    2026-02-15T10:30:00Z
Content:    User prefers TypeScript for frontend
Supersedes: mem_old456
Topics:     [typescript, frontend, preferences]
```

### JSON Output

Support `--json` flag for machine-readable output.

## Testing

- Test all CRUD operations
- Test filtering by scope/category
- Test semantic search
- Test JSON output format
