---
id: ase-u200
status: closed
deps: [ase-w2n6]
links: []
created: 2026-02-09T03:25:42Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-qnjh
---
# Define daemon RPC protocol

## Background

The CLI communicates with the daemon via Unix socket using JSON-RPC. As we add Matrix, flows, and agent creation, we need a well-defined RPC protocol.

## Why This Matters

Without a clear protocol:
- CLI and daemon implementations diverge
- Error handling is inconsistent
- Adding new features requires ad-hoc changes
- Testing is harder

## Implementation Details

### RPC Methods

| Method | Description |
|--------|-------------|
| **Service** ||
| service.status | Get daemon and subsystem status |
| service.shutdown | Gracefully stop daemon |
| **Matrix** ||
| matrix.status | Get Matrix connection status |
| matrix.rooms.list | List rooms, optionally filter |
| matrix.rooms.create | Create a new room |
| matrix.rooms.members | Get room members |
| matrix.send | Send message to room |
| matrix.history | Get message history |
| **Flows** ||
| flows.list | List registered flows |
| flows.run | Execute a flow |
| flows.status | Get running flow status |
| flows.cancel | Cancel running flow |
| **Triggers** ||
| triggers.list | List active triggers |
| triggers.add | Register new trigger |
| triggers.remove | Remove trigger |
| triggers.fire | Manually fire trigger |
| triggers.stats | Get trigger statistics |
| **Agents** ||
| agents.create | Create new agent (for @ayo) |
| agents.refine | Refine agent prompt |
| agents.invoke | Invoke agent with message |
| agents.capabilities | Get/refresh capabilities |

### Request/Response Format

```go
// internal/daemon/protocol/types.go

type Request struct {
    JSONRPC string      `json:"jsonrpc"` // "2.0"
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
    ID      int         `json:"id"`
}

type Response struct {
    JSONRPC string      `json:"jsonrpc"`
    Result  interface{} `json:"result,omitempty"`
    Error   *Error      `json:"error,omitempty"`
    ID      int         `json:"id"`
}

type Error struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}

// Standard error codes
const (
    ErrCodeParse       = -32700
    ErrCodeInvalidReq  = -32600
    ErrCodeMethodNotFound = -32601
    ErrCodeInvalidParams = -32602
    ErrCodeInternal    = -32603
    
    // Application errors (positive codes)
    ErrCodeAgentNotFound = 1001
    ErrCodeFlowNotFound  = 1002
    ErrCodeRoomNotFound  = 1003
    ErrCodeTimeout       = 1004
    ErrCodeUnauthorized  = 1005
)
```

### Example Methods

```go
// Matrix send
type MatrixSendParams struct {
    Room    string `json:"room"`
    Content string `json:"content"`
    AsAgent string `json:"as_agent,omitempty"`
}

type MatrixSendResult struct {
    EventID string `json:"event_id"`
}

// Flow run
type FlowRunParams struct {
    Name   string         `json:"name"`
    Params map[string]any `json:"params,omitempty"`
    Async  bool           `json:"async,omitempty"`
}

type FlowRunResult struct {
    RunID  string                 `json:"run_id"`
    Status string                 `json:"status"`
    Steps  map[string]StepResult `json:"steps,omitempty"`
}
```

### Files to Create

1. `internal/daemon/protocol/types.go` - All type definitions
2. `internal/daemon/protocol/errors.go` - Error codes and helpers
3. `internal/daemon/protocol/methods.go` - Method name constants
4. Update `internal/daemon/server.go` - Use protocol types
5. Update CLI commands to use protocol types

## Acceptance Criteria

- [ ] All RPC methods documented with params/results
- [ ] Standard JSON-RPC 2.0 format
- [ ] Application-specific error codes defined
- [ ] Protocol types in dedicated package
- [ ] CLI and daemon share protocol types
- [ ] Unit tests for serialization/deserialization

