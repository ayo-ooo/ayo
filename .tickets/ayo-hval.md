---
id: ayo-hval
status: open
deps: [ayo-hscm]
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-hitl
tags: [human-in-the-loop, validation]
---
# Task: Input Validation and Re-prompting

## Summary

Implement validation logic for input responses and re-prompting for invalid input. Invalid input should never cause failure - it should trigger a helpful re-prompt.

## Validation Rules

### Required Fields
```go
if field.Required && isEmpty(value) {
    return ValidationError{
        Field:   field.Name,
        Message: fmt.Sprintf("%s is required", field.Label),
    }
}
```

### Type Validation

| Type | Validation |
|------|------------|
| `text` | minLength, maxLength, pattern |
| `number` | min, max, integer vs float |
| `select` | value in options |
| `multiselect` | all values in options |
| `date` | parseable as date |
| `email` | valid email format |
| `url` | valid URL format |

### Custom Validation

```json
{
  "name": "amount",
  "type": "number",
  "validation": {
    "min": 0,
    "max": 10000,
    "message": "Amount must be between $0 and $10,000"
  }
}
```

## Re-prompting

When validation fails, the renderer should:

1. Show the error message
2. Preserve previous valid values
3. Re-prompt only the invalid field
4. Allow retry (up to configurable limit)

### CLI Re-prompt
```
┌ Error ───────────────────────────────────────────────────────┐
│ Amount must be between $0 and $10,000                        │
└──────────────────────────────────────────────────────────────┘

Amount: _
```

### Chat Re-prompt
```
Agent: That amount is too high. Please enter a value between $0 and $10,000.

User: 5000

Agent: Got it - $5,000.
```

## Implementation

```go
func (v *Validator) Validate(field Field, value any) error {
    if field.Required && isEmpty(value) {
        return &ValidationError{...}
    }
    
    switch field.Type {
    case "number":
        return v.validateNumber(field, value)
    case "text":
        return v.validateText(field, value)
    // ...
    }
}
```

## Files to Create

- `internal/hitl/validation.go` - Validation logic
- `internal/hitl/validation_test.go` - Tests

## Acceptance Criteria

- [ ] All field types have validation
- [ ] Custom validation rules work
- [ ] Error messages are helpful
- [ ] Re-prompting preserves context
- [ ] Max retry limit enforced
- [ ] Final failure returns error
