---
id: am-rvt0
status: open
deps: [am-q2ni]
links: []
created: 2026-02-18T03:15:03Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-8epk
---
# Add input validation on squad dispatch

Validate input against squad's input.jsonschema when dispatching work to a squad.

## Context
- Squad dispatch will happen via CLI or @ayo delegation
- Validation should reject at boundary if schema defined and input invalid

## Implementation
```go
// internal/squads/dispatch.go (new file)

type DispatchInput struct {
    Prompt string         // Free-form prompt
    Data   map[string]any // Structured data (validated if schema exists)
}

func (s *Squad) ValidateInput(input DispatchInput) error {
    if s.Schemas == nil || s.Schemas.Input == nil {
        return nil // Free-form mode
    }
    return s.Schemas.Input.Validate(input.Data)
}

func (s *Squad) Dispatch(ctx context.Context, input DispatchInput) (*DispatchResult, error) {
    if err := s.ValidateInput(input); err != nil {
        return nil, fmt.Errorf("input validation failed: %w", err)
    }
    // ... proceed with dispatch
}
```

## Files to Create
- internal/squads/dispatch.go

## Dependencies
- am-q2ni (schema loading)

## Acceptance
- Input validated if schema exists
- Clear error message on validation failure
- Free-form input accepted if no schema

