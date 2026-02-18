---
id: am-2wpf
status: closed
deps: [am-r6do]
links: []
created: 2026-02-18T03:18:48Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-okzf
---
# Implement AgentInvoker with sandbox awareness

Implement the AgentInvoker interface that can invoke agents in sandboxes.

## Context
- AgentInvoker interface exists but is unimplemented
- Needs to invoke agents with proper sandbox context
- Location: internal/flows/yaml_executor.go

## Implementation
```go
// internal/flows/invoker.go (new file)

type SandboxAwareInvoker struct {
    daemonClient *daemon.Client
}

func (i *SandboxAwareInvoker) Invoke(ctx context.Context, agent, prompt string) (string, error) {
    // Parse agent handle
    handle := agent
    
    // Invoke via daemon (which handles sandbox)
    resp, err := i.daemonClient.InvokeAgent(ctx, daemon.InvokeRequest{
        Agent:  handle,
        Prompt: prompt,
    })
    if err != nil {
        return "", err
    }
    
    return resp.Output, nil
}

func NewSandboxAwareInvoker(client *daemon.Client) flows.AgentInvoker {
    return &SandboxAwareInvoker{daemonClient: client}
}
```

## Files to Create
- internal/flows/invoker.go

## Files to Modify
- internal/daemon/flow_rpc.go (wire up invoker)

## Acceptance
- AgentInvoker implemented
- Agents invoked via daemon
- Sandbox context applied

