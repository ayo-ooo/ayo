---
id: am-vego
status: closed
deps: [am-rvt0]
links: []
created: 2026-02-18T03:18:20Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-hin9
---
# Add squad.dispatch RPC to daemon

Add RPC endpoint for dispatching work to a squad.

## Context
- CLI communicates with daemon via RPC
- Need squad.dispatch RPC for synchronous invocation
- Returns result when squad completes

## Implementation
```go
// internal/daemon/squad_dispatch_rpc.go (new file)

type SquadDispatchRequest struct {
    SquadName string         `json:"squad_name"`
    Prompt    string         `json:"prompt"`
    Data      map[string]any `json:"data,omitempty"`
}

type SquadDispatchResponse struct {
    Output map[string]any `json:"output,omitempty"`
    Raw    string         `json:"raw,omitempty"`
    Error  string         `json:"error,omitempty"`
}

func (h *RPCHandler) SquadDispatch(ctx context.Context, req SquadDispatchRequest) (*SquadDispatchResponse, error)
```

## Files to Create
- internal/daemon/squad_dispatch_rpc.go

## Dependencies
- am-rvt0 (dispatch infrastructure)

## Acceptance
- RPC endpoint registered
- Validates input against schema
- Returns result synchronously

