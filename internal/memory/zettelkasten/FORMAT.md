# Zettelkasten Memory File Format

This document defines the file format for memories stored in the Zettelkasten memory provider.

## Overview

Each memory is stored as a Markdown file with TOML frontmatter. The frontmatter contains metadata, and the body contains the memory content.

## Directory Structure

```
~/.local/share/ayo/memory/
├── index.md              # Auto-generated overview (optional)
├── topics/               # Memories organized by topic (symlinks)
│   ├── go-preferences/
│   │   └── mem_01HX...md -> ../../facts/mem_01HX...md
│   └── testing/
│       └── mem_01HY...md -> ../../preferences/mem_01HY...md
├── facts/                # Category: factual information
│   └── mem_01HX...md
├── preferences/          # Category: user preferences
│   └── mem_01HY...md
├── corrections/          # Category: corrections to agent behavior
│   └── mem_01HZ...md
├── patterns/             # Category: observed behavioral patterns
│   └── mem_01HA...md
└── .index.sqlite         # Derived index with embeddings and FTS (rebuildable)
```

## File Naming

Files are named with the memory ID and `.md` extension:
- Pattern: `{id}.md`
- Example: `mem_01HX9Z7KQPRT5J8YCNM6WVF4GT.md`

IDs use the ULID format (Universally Unique Lexicographically Sortable Identifier):
- 26 characters, base32 encoded
- Prefix: `mem_`
- Sortable by creation time

## File Format

```markdown
+++
id = "mem_01HX9Z7KQPRT5J8YCNM6WVF4GT"
created = 2024-01-15T10:30:00Z
updated = 2024-01-15T10:30:00Z
category = "preference"
status = "active"
topics = ["go", "testing", "style"]
confidence = 0.85

[source]
session_id = "ses_01HX9Z6ABC..."
message_id = "msg_01HX9Z7DEF..."

[scope]
agent = ""          # Empty for global memories
path = ""           # Empty for non-path-scoped memories

[access]
last_accessed = 2024-01-20T15:45:00Z
access_count = 5

[supersession]
supersedes = ""     # ID of memory this one replaces
superseded_by = ""  # ID of memory that replaced this one
reason = ""         # Why the supersession occurred

[links]
related = ["mem_01HX...", "mem_01HY..."]  # Related memories

[unclear]
flagged = false
reason = ""         # Why the memory needs clarification
+++

User prefers table-driven tests in Go with descriptive test names.
They find this style more readable than individual test functions.
```

## Field Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique memory identifier (ULID with `mem_` prefix) |
| `created` | datetime | When the memory was created (RFC 3339) |
| `category` | string | One of: `preference`, `fact`, `correction`, `pattern` |
| `status` | string | One of: `active`, `superseded`, `archived`, `forgotten` |
| `topics` | string[] | List of topic tags |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `updated` | datetime | `created` | When the memory was last modified |
| `confidence` | float | 1.0 | Confidence score (0.0-1.0) |

### Source Section

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | ID of the session where this memory was formed |
| `message_id` | string | ID of the specific message that triggered this memory |

### Scope Section

| Field | Type | Description |
|-------|------|-------------|
| `agent` | string | Agent handle this memory applies to (empty = global) |
| `path` | string | Directory path this memory is scoped to (empty = all paths) |

### Access Section

| Field | Type | Description |
|-------|------|-------------|
| `last_accessed` | datetime | When this memory was last retrieved |
| `access_count` | int | Number of times this memory has been retrieved |

### Supersession Section

| Field | Type | Description |
|-------|------|-------------|
| `supersedes` | string | ID of memory this one replaces |
| `superseded_by` | string | ID of memory that replaced this one |
| `reason` | string | Reason for the supersession |

### Links Section

| Field | Type | Description |
|-------|------|-------------|
| `related` | string[] | IDs of related memories |

### Unclear Section

| Field | Type | Description |
|-------|------|-------------|
| `flagged` | bool | True if memory needs clarification |
| `reason` | string | Why clarification is needed |

## Categories

| Category | Description | Examples |
|----------|-------------|----------|
| `preference` | User preferences and choices | "Prefers Go over Python", "Uses VS Code" |
| `fact` | Factual information | "API key is in ~/.config/", "Project uses PostgreSQL" |
| `correction` | Corrections to agent behavior | "Don't use Cobra for this project" |
| `pattern` | Observed behavioral patterns | "Usually works on backend first" |

## Status Lifecycle

```
                    ┌─────────────┐
        create      │   active    │
       ─────────►   │             │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
  │ superseded  │  │  archived   │  │  forgotten  │
  │             │  │             │  │             │
  └─────────────┘  └─────────────┘  └─────────────┘
       │
       ▼
  (new memory)
```

- **active**: Normal state, memory is in use
- **superseded**: Replaced by a newer memory (soft-linked)
- **archived**: Manually archived, not used but preserved
- **forgotten**: Soft-deleted, will be pruned

## Index File (.index.sqlite)

The SQLite index is a derived file that can be rebuilt from the markdown files.

### Schema

```sql
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    category TEXT NOT NULL,
    status TEXT NOT NULL,
    agent_handle TEXT,
    path_scope TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    confidence REAL DEFAULT 1.0,
    embedding BLOB,  -- float32 vector, nomic-embed-text dimensions
    file_path TEXT NOT NULL,
    file_mtime INTEGER NOT NULL
);

CREATE TABLE topics (
    memory_id TEXT NOT NULL,
    topic TEXT NOT NULL,
    PRIMARY KEY (memory_id, topic),
    FOREIGN KEY (memory_id) REFERENCES memories(id)
);

CREATE TABLE links (
    source_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    PRIMARY KEY (source_id, target_id),
    FOREIGN KEY (source_id) REFERENCES memories(id),
    FOREIGN KEY (target_id) REFERENCES memories(id)
);

-- Full-text search
CREATE VIRTUAL TABLE memories_fts USING fts5(
    content,
    content='memories',
    content_rowid='rowid'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER memories_ai AFTER INSERT ON memories BEGIN
    INSERT INTO memories_fts(rowid, content) VALUES (new.rowid, new.content);
END;

CREATE TRIGGER memories_ad AFTER DELETE ON memories BEGIN
    INSERT INTO memories_fts(memories_fts, rowid, content) VALUES('delete', old.rowid, old.content);
END;

CREATE TRIGGER memories_au AFTER UPDATE ON memories BEGIN
    INSERT INTO memories_fts(memories_fts, rowid, content) VALUES('delete', old.rowid, old.content);
    INSERT INTO memories_fts(rowid, content) VALUES (new.rowid, new.content);
END;
```

## Conflict Resolution

When multiple sessions modify memories concurrently:

1. **Auto-merge**: If changes are to different fields, merge automatically
2. **Unclear flag**: If changes conflict, set `unclear.flagged = true` with reason
3. **Agent asks**: The next agent interaction naturally asks for clarification

## File Examples

### Preference Memory

```markdown
+++
id = "mem_01HX9Z7KQPRT5J8YCNM6WVF4GT"
created = 2024-01-15T10:30:00Z
category = "preference"
status = "active"
topics = ["go", "testing"]
confidence = 0.9

[source]
session_id = "ses_01HX9Z6ABC"
+++

User prefers table-driven tests in Go.
```

### Correction Memory

```markdown
+++
id = "mem_01HY8A6BMNST4K7XDPL2QRF5HU"
created = 2024-01-16T14:20:00Z
category = "correction"
status = "active"
topics = ["project", "dependencies"]
confidence = 1.0

[source]
session_id = "ses_01HY8A5CDE"
message_id = "msg_01HY8A6FGH"

[scope]
path = "/Users/alex/Code/myproject"
+++

Do NOT use Cobra for CLI in this project. Use standard library flag package instead.
User explicitly corrected this when I suggested Cobra.
```

### Superseded Memory

```markdown
+++
id = "mem_01HX9Z7KQPRT5J8YCNM6WVF4GT"
created = 2024-01-15T10:30:00Z
updated = 2024-01-17T09:00:00Z
category = "preference"
status = "superseded"
topics = ["editor"]

[supersession]
superseded_by = "mem_01HZ7B8CLNRT6K9YEPM3WSG5IV"
reason = "User switched from VS Code to Neovim"
+++

User prefers VS Code for development.
```
