---
id: am-ncn5
status: open
deps: [am-q2ni, am-rvt0]
links: []
created: 2026-02-18T03:15:08Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-8epk
---
# Add output validation on squad completion

Validate output against squad's output.jsonschema when squad completes work.

## Context
- Squad output should be validated before returning to caller
- This ensures squads honor their contract

## Implementation
```go
// Add to internal/squads/dispatch.go

type DispatchResult struct {
    Output map[string]any
    Raw    string // For free-form output
}

func (s *Squad) ValidateOutput(result *DispatchResult) error {
    if s.Schemas == nil || s.Schemas.Output == nil {
        return nil // Free-form mode
    }
    return s.Schemas.Output.Validate(result.Output)
}
```

## Files to Modify
- internal/squads/dispatch.go

## Dependencies
- am-q2ni (schema loading)
- am-rvt0 (dispatch infrastructure)

## Acceptance
- Output validated if schema exists
- Clear error message on validation failure
- Free-form output accepted if no schema

