---
id: am-r2lh
status: closed
deps: [am-q2ni, am-d2rd]
links: []
created: 2026-02-18T03:19:12Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-okzf
---
# Validate flow I/O against squad schemas

Validate flow step I/O against target squad's schemas.

## Context
- If a flow step targets a squad with schemas, validate I/O
- Ensures type safety in cross-squad communication

## Implementation
```go
// internal/flows/validation.go (new file)

func (e *YAMLExecutor) validateStepIO(ctx context.Context, step Step, input, output string) error {
    if step.Squad == "" {
        return nil // No squad, no schema validation
    }
    
    squad, err := e.squadService.Get(step.Squad)
    if err != nil {
        return err
    }
    
    // Validate input
    if squad.Schemas.Input != nil {
        if err := squad.Schemas.Input.ValidateString(input); err != nil {
            return fmt.Errorf("input validation failed: %w", err)
        }
    }
    
    // Validate output
    if squad.Schemas.Output != nil {
        if err := squad.Schemas.Output.ValidateString(output); err != nil {
            return fmt.Errorf("output validation failed: %w", err)
        }
    }
    
    return nil
}
```

## Files to Create
- internal/flows/validation.go

## Dependencies
- am-q2ni (squad schema loading)
- am-d2rd (squad context in invoker)

## Acceptance
- Input validated before squad dispatch
- Output validated after squad completion
- Clear error messages on validation failure

