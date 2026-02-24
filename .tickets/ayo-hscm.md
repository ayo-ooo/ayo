---
id: ayo-hscm
status: open
deps: []
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-hitl
tags: [human-in-the-loop, schema, design]
---
# Task: Define Input Request Schema

## Summary

Design and implement the JSON schema for agent input requests. This schema must be expressive enough to represent complex forms while remaining simple enough to degrade gracefully to conversational Q&A.

## Schema Requirements

### Field Types
- `text` - Single line text input
- `textarea` - Multi-line text input
- `select` - Single choice from options
- `multiselect` - Multiple choices from options
- `confirm` - Yes/No boolean
- `number` - Numeric input with optional min/max
- `date` - Date input with natural language parsing
- `file` - File upload/attachment

### Validation Rules
- `required` - Field must have value
- `minLength` / `maxLength` - Text length bounds
- `min` / `max` - Numeric bounds
- `pattern` - Regex validation
- `options` - Enum validation for select types

### Metadata
- `timeout` - How long to wait for response
- `recipient` - Who to ask (owner, email, chat)
- `persona` - How agent should present itself
- `context` - Background for the request

## Implementation

Create `internal/hitl/schema.go`:

```go
type InputRequest struct {
    ID        string        `json:"id"`
    Timeout   time.Duration `json:"timeout"`
    Recipient Recipient     `json:"recipient"`
    Context   string        `json:"context"`
    Fields    []Field       `json:"fields"`
    Persona   *Persona      `json:"persona,omitempty"`
}

type Field struct {
    Name        string      `json:"name"`
    Type        FieldType   `json:"type"`
    Label       string      `json:"label"`
    Description string      `json:"description,omitempty"`
    Required    bool        `json:"required"`
    Default     any         `json:"default,omitempty"`
    Options     []Option    `json:"options,omitempty"`
    Validation  *Validation `json:"validation,omitempty"`
}

type Recipient struct {
    Type    string `json:"type"` // owner, email, chat
    Address string `json:"address,omitempty"`
}

type Persona struct {
    Name      string `json:"name"`
    Signature string `json:"signature,omitempty"`
}
```

## Files to Create

- `internal/hitl/schema.go` - Schema types
- `internal/hitl/schema_test.go` - Schema tests
- `internal/hitl/validate.go` - Schema validation
- `internal/hitl/validate_test.go` - Validation tests

## Acceptance Criteria

- [ ] Schema can represent all field types
- [ ] Schema validates correctly
- [ ] JSON serialization/deserialization works
- [ ] Validation rules enforce constraints
- [ ] Schema is documented with examples
