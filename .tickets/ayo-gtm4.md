---
id: ayo-gtm4
status: deferred
deps: []
links: []
created: 2026-02-24T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-gtm1
tags: [gtm, refactoring]
---
# Task: Consolidate AgentInvoker Interface

## Summary

The `AgentInvoker` interface is defined in 9+ locations across the codebase. Consolidate to a single canonical definition.

## Current Locations

```
internal/daemon/invoker.go
internal/daemon/invoker_test.go
internal/daemon/server.go
internal/daemon/squad_invoker.go
internal/daemon/squad_rpc.go
internal/run/fantasy_tools.go
internal/run/run.go
internal/squads/invoker.go
internal/squads/service.go
```

## Analysis Needed

1. Determine which definition is canonical
2. Check if definitions are identical or have variations
3. Identify if some are type aliases vs actual definitions
4. Check import cycles that may have forced duplication

## Proposed Solution

Create `internal/interfaces/invoker.go`:

```go
package interfaces

import (
    "context"
)

// AgentInvoker defines the interface for invoking an agent.
type AgentInvoker interface {
    Invoke(ctx context.Context, req *InvokeRequest) (*InvokeResponse, error)
}

// InvokeRequest contains the parameters for invoking an agent.
type InvokeRequest struct {
    AgentHandle string
    Prompt      string
    SessionID   string
    // ... other fields
}

// InvokeResponse contains the result of invoking an agent.
type InvokeResponse struct {
    Output    string
    ToolCalls []ToolCallResult
    // ... other fields
}
```

## Implementation Steps

1. [ ] Audit all 9 locations to understand variations
2. [ ] Create `internal/interfaces/` package
3. [ ] Define canonical `AgentInvoker` interface
4. [ ] Update all imports to use `interfaces.AgentInvoker`
5. [ ] Remove duplicate definitions
6. [ ] Run tests to verify no breakage
7. [ ] Check for import cycle issues

## Acceptance Criteria

- [ ] AgentInvoker defined in exactly one location
- [ ] All packages import from canonical location
- [ ] No circular import errors
- [ ] All tests pass
