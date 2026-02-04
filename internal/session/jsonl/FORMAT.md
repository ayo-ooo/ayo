# Session JSONL File Format

## Overview

Session files are stored as JSONL (JSON Lines) format where each line is a complete JSON object. This format provides:
- Append-only writes (efficient for real-time streaming)
- Easy parsing (line-by-line)
- Human-readable
- Supports partial reads (for large sessions)

## Directory Structure

```
~/.local/share/ayo/sessions/
├── index.sqlite           # Derived index for search/listing (rebuildable)
├── @ayo/                   # Sessions by agent
│   ├── 2026-01/           # By year-month
│   │   ├── sess-abc123.jsonl
│   │   └── sess-def456.jsonl
│   └── 2026-02/
│       └── sess-ghi789.jsonl
└── @custom-agent/
    └── 2026-01/
        └── sess-xyz000.jsonl
```

## File Format

### Line 0: Session Metadata (Header)

```json
{
  "type": "session",
  "id": "sess-abc123",
  "agent_handle": "@ayo",
  "title": "Debug authentication issue",
  "source": "ayo",
  "created_at": "2026-01-15T10:30:00Z",
  "updated_at": "2026-01-15T11:45:00Z",
  "finished_at": "2026-01-15T11:45:30Z",
  "chain_depth": 0,
  "chain_source": "",
  "input_schema": null,
  "output_schema": null,
  "structured_input": null,
  "structured_output": null,
  "message_count": 12
}
```

### Lines 1+: Messages

Each subsequent line is a message:

```json
{
  "type": "message",
  "id": "msg-001",
  "role": "user",
  "model": null,
  "provider": null,
  "created_at": "2026-01-15T10:30:01Z",
  "updated_at": "2026-01-15T10:30:01Z",
  "finished_at": "2026-01-15T10:30:01Z",
  "parts": [
    {"type": "text", "data": {"text": "Help me debug the auth issue"}}
  ]
}
```

### Content Part Types

#### Text Content
```json
{"type": "text", "data": {"text": "The response text..."}}
```

#### Reasoning/Thinking Content
```json
{
  "type": "reasoning",
  "data": {
    "text": "Let me think about this...",
    "signature": "thinking",
    "started_at": "2026-01-15T10:30:02Z",
    "finished_at": "2026-01-15T10:30:05Z"
  }
}
```

#### Tool Call
```json
{
  "type": "tool_call",
  "data": {
    "id": "call_abc123",
    "name": "bash",
    "input": {"command": "ls -la", "description": "List files"},
    "provider_executed": true,
    "finished": true,
    "started_at": "2026-01-15T10:30:10Z",
    "finished_at": "2026-01-15T10:30:12Z"
  }
}
```

#### Tool Result
```json
{
  "type": "tool_result",
  "data": {
    "tool_call_id": "call_abc123",
    "name": "bash",
    "content": [{"type": "text", "text": "file1.go\nfile2.go"}],
    "is_error": false
  }
}
```

#### File Content
```json
{
  "type": "file",
  "data": {
    "filename": "screenshot.png",
    "data": "base64...",
    "media_type": "image/png"
  }
}
```

#### Finish Marker
```json
{
  "type": "finish",
  "data": {
    "reason": "end_turn",
    "time": "2026-01-15T10:30:30Z",
    "message": null
  }
}
```

## Index Schema (SQLite)

The index is derived and rebuildable from source files:

```sql
CREATE TABLE session_index (
    id TEXT PRIMARY KEY,
    agent_handle TEXT NOT NULL,
    title TEXT NOT NULL,
    source TEXT NOT NULL,
    file_path TEXT NOT NULL,       -- Path to JSONL file
    message_count INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    finished_at INTEGER,
    chain_depth INTEGER DEFAULT 0
);

CREATE INDEX idx_session_agent ON session_index(agent_handle);
CREATE INDEX idx_session_created ON session_index(created_at DESC);

-- FTS for title search
CREATE VIRTUAL TABLE session_fts USING fts5(title, content=session_index, content_rowid=rowid);
```

## Implementation Notes

### Writing Sessions
1. Create file on session start
2. Write header line
3. Append messages as they arrive (streaming-friendly)
4. Update header's `updated_at` and `message_count` on close

### Reading Sessions
1. Read first line for metadata
2. Stream remaining lines for messages
3. Support offset/limit for pagination

### Indexing
1. Scan all JSONL files
2. Read only header line (first line)
3. Insert into SQLite index
4. Rebuild on `ayo sessions reindex`

### Migration from SQLite
1. Read sessions and messages from SQLite
2. Write JSONL files in proper structure
3. Skip files that already exist (unless --overwrite)
