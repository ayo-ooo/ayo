---
id: ayo-dicu
status: closed
deps: [ayo-kkxg]
links: []
created: 2026-02-23T22:15:27Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-whmn
tags: [tools, filesystem]
---
# Implement file_request tool

Create tool for agents to request modifications to host files.

## Tool Specification

```json
{
  "name": "file_request",
  "description": "Request permission to modify a file on the host system",
  "parameters": {
    "action": {
      "type": "string",
      "enum": ["create", "update", "delete"],
      "description": "Type of file operation"
    },
    "path": {
      "type": "string",
      "description": "Path relative to host home (e.g., 'Projects/app/main.go')"
    },
    "content": {
      "type": "string",
      "description": "File content (for create/update)"
    },
    "reason": {
      "type": "string",
      "description": "Why this change is needed (shown to user)"
    }
  },
  "required": ["action", "path", "reason"]
}
```

## Implementation

### Location
`internal/tools/file_request.go`

### Flow

1. Agent calls `file_request(action, path, content, reason)`
2. Tool validates path is within allowed boundaries
3. Tool sends request to daemon via RPC
4. Daemon checks auto-approval settings (--no-jodas, per-agent, session cache)
5. If not auto-approved, daemon sends approval request to terminal
6. Terminal displays approval UI (ayo-c5mt)
7. On approval, daemon writes file to host
8. Tool returns result to agent

### Response

```json
{
  "status": "approved|denied|error",
  "path": "/Users/alex/Projects/app/main.go",
  "message": "File updated successfully"
}
```

## Files to Modify

- Create `internal/tools/file_request.go`
- Update `internal/tools/registry.go` to register tool
- Add RPC method to `internal/daemon/rpc.go`
- Add approval handler interface

## Testing

- Unit tests for path validation
- Integration test with mock approval
- Test auto-approval modes
