---
id: ayo-vclt
status: closed
deps: [ayo-dicu]
links: []
created: 2026-02-23T23:13:32Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-whmn
tags: [security, logging]
---
# Implement file modification audit logging

Log all file modifications for security and accountability.

## Log Location

`~/.local/share/ayo/audit.log`

## Log Format

JSON Lines format for easy parsing:

```json
{"ts":"2026-02-23T15:30:00Z","agent":"@ayo","session":"abc123","action":"update","path":"/Users/alex/Projects/app/main.go","approval":"no_jodas","size":4096,"hash":"sha256:abc..."}
{"ts":"2026-02-23T15:31:00Z","agent":"@crush","session":"abc123","action":"create","path":"/Users/alex/Projects/app/util.go","approval":"user_approved","size":1234,"hash":"sha256:def..."}
```

## Fields

| Field | Type | Description |
|-------|------|-------------|
| `ts` | string | ISO 8601 timestamp |
| `agent` | string | Agent that made the modification |
| `session` | string | Session ID for grouping |
| `action` | string | "create", "update", "delete" |
| `path` | string | Absolute path on host |
| `approval` | string | How approval was obtained |
| `size` | int | File size in bytes (for create/update) |
| `hash` | string | SHA256 of new content (for verification) |

## Approval Types

- `user_approved` - User pressed Y in prompt
- `session_cache` - User pressed A earlier in session  
- `no_jodas` - --no-jodas flag was used
- `agent_config` - Agent has auto_approve: true
- `global_config` - Global no_jodas setting

## Implementation

### Location
`internal/audit/audit.go`

### Interface

```go
type AuditLogger interface {
    Log(entry AuditEntry) error
    Query(filter AuditFilter) ([]AuditEntry, error)
}

type AuditEntry struct {
    Timestamp   time.Time `json:"ts"`
    Agent       string    `json:"agent"`
    Session     string    `json:"session"`
    Action      string    `json:"action"`
    Path        string    `json:"path"`
    Approval    string    `json:"approval"`
    Size        int64     `json:"size,omitempty"`
    ContentHash string    `json:"hash,omitempty"`
}
```

### CLI Command

```bash
# View recent audit entries
ayo audit list

# View entries for specific agent
ayo audit list --agent @ayo

# View entries for session
ayo audit list --session abc123

# Export to CSV
ayo audit export --format csv > audit.csv
```

## Files to Create

- `internal/audit/audit.go` - Logger implementation
- `internal/audit/audit_test.go` - Tests
- `cmd/ayo/audit.go` - CLI commands

## Log Rotation

- Keep last 10MB of logs
- Rotate when size exceeded
- Keep 3 rotated files (audit.log.1, .2, .3)

## Testing

- Test log file creation
- Test JSON format validity
- Test rotation behavior
- Test concurrent writes
