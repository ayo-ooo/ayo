---
id: am-r6do
status: open
deps: [am-xpxb]
links: []
created: 2026-02-18T03:20:06Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-11v2
---
# Add agent invoke RPC to daemon

Add RPC endpoint for invoking an agent directly (not in a squad).

## Context
- Flows need to invoke agents via daemon
- Direct agent invocation also uses this
- Agent runs in @ayo sandbox with home directory

## Implementation
```go
// internal/daemon/agent_invoke_rpc.go (new file)

type AgentInvokeRequest struct {
    Agent  string `json:"agent"`
    Prompt string `json:"prompt"`
}

type AgentInvokeResponse struct {
    Output string `json:"output"`
    Error  string `json:"error,omitempty"`
}

func (h *RPCHandler) AgentInvoke(ctx context.Context, req AgentInvokeRequest) (*AgentInvokeResponse, error)
```

## Files to Create
- internal/daemon/agent_invoke_rpc.go

## Dependencies
- am-xpxb (agent home mounting)

## Acceptance
- RPC endpoint registered
- Agent invoked in @ayo sandbox
- Home directory mounted

