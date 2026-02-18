---
id: am-d2rd
status: open
deps: [am-2wpf, am-mh6x]
links: []
created: 2026-02-18T03:19:02Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-okzf
---
# Extend AgentInvoker to support squad context

Modify AgentInvoker to accept squad context for cross-squad invocation.

## Context
- When step has squad field, invoke in that squad's sandbox
- AgentInvoker needs squad parameter

## Implementation
```go
// internal/flows/yaml_executor.go

// Extend interface
type AgentInvoker interface {
    Invoke(ctx context.Context, agent, prompt string) (string, error)
    InvokeInSquad(ctx context.Context, squad, agent, prompt string) (string, error)
}

// Update step execution
func (e *YAMLExecutor) executeAgentStep(ctx context.Context, step Step) (*StepResult, error) {
    if step.Squad != "" {
        return e.AgentInvoker.InvokeInSquad(ctx, step.Squad, step.Agent, prompt)
    }
    return e.AgentInvoker.Invoke(ctx, step.Agent, prompt)
}
```

## Files to Modify
- internal/flows/yaml_executor.go
- internal/flows/invoker.go

## Dependencies
- am-2wpf (base invoker implementation)
- am-mh6x (squad field in spec)

## Acceptance
- InvokeInSquad method added
- Step execution uses squad context
- Squad started if not running

