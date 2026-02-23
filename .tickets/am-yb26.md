---
id: am-yb26
status: closed
deps: []
links: []
created: 2026-02-20T02:50:40Z
type: task
priority: 2
assignee: Alex Cabrera
tags: [squads, schemas]
---
# Validate squad output against output.jsonschema after agent response

After an agent completes work in a squad, the output should be validated against output.jsonschema if present. This ensures squad responses conform to expected formats.

## Current State

### Already Implemented

| Component | File | Lines |
|-----------|------|-------|
| `ValidateOutput()` method | internal/squads/dispatch.go | 79-97 |
| `loadSchemaFile()` for output.jsonschema | internal/squads/schema.go | 70-84 |
| Schema loading in service | internal/squads/service.go | 121-124, 165-168 |

### ValidateOutput Implementation (dispatch.go:79-97)

```go
func (s *Squad) ValidateOutput(result *DispatchResult) error {
    if s.Schemas == nil || s.Schemas.Output == nil {
        return nil
    }
    if result.Output == nil && result.Raw != "" {
        return nil
    }
    if result.Output != nil {
        if err := schema.ValidateAgainstSchema(result.Output, *s.Schemas.Output); err != nil {
            return &ValidationError{Direction: "output", Err: err}
        }
    }
    return nil
}
```

## The Gap

`ValidateOutput()` exists but is **never called** in the dispatch flow.

In dispatch.go:164-167, the result is returned directly without validation:
```go
return &DispatchResult{
    Raw:      result.Response,
    RoutedTo: targetAgent,
}, nil
```

**Note:** Output validation IS called in `HandleSquadDispatch` RPC (squad_rpc.go:497-503):
```go
if err := squad.ValidateOutput(result); err != nil {
    return SquadDispatchResult{
        Output: result.Output,
        Raw:    result.Raw,
        Error:  "output validation failed: " + err.Error(),
    }, nil
}
```

However, this returns an error field rather than logging a warning as specified.

## Implementation Plan

### Option 1: Log warning and continue (per acceptance criteria)

Modify squad_rpc.go:497-503:
```go
if err := squad.ValidateOutput(result); err != nil {
    debug.Log("output validation failed", "squad", name, "error", err)
    // Continue with response, don't fail
}
```

### Option 2: Also validate in dispatch.go for non-RPC callers

Add validation in dispatch.go after line 167:
```go
// Validate output (warning only)
if err := s.ValidateOutput(dispatchResult); err != nil {
    debug.Log("output validation warning", "error", err)
}
return dispatchResult, nil
```

## Key Code Locations

| File | Lines | Purpose |
|------|-------|---------|
| internal/squads/dispatch.go | 79-97 | `ValidateOutput()` method |
| internal/squads/dispatch.go | 164-167 | Result returned without validation |
| internal/daemon/squad_rpc.go | 497-503 | Validation called but returns error |

## Acceptance Criteria

- ⬜ Agent output is validated against output.jsonschema (partially: only in RPC path)
- ⬜ Validation failures are logged as warnings (currently returns error)
- ✅ Valid output is returned to caller normally
