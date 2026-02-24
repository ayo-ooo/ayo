---
id: ayo-oxj6
status: closed
deps: [ayo-n88v]
links: []
created: 2026-02-23T22:15:47Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-xfu3
tags: [squads, schemas]
---
# Add I/O schema enforcement for squads

Validate dispatched input against JSON Schema if present in squad. Validate output before returning. Return clear validation errors.

## Context

Squads can define strict input/output contracts using JSON Schema. This enables:
- Clear API contracts for squad capabilities
- Validation of user input before processing
- Validation of agent output before returning
- Better error messages for malformed requests

## Schema Files

```
squad-dir/
├── ayo.json
├── SQUAD.md
├── schemas/
│   ├── input.json     # Input validation
│   └── output.json    # Output validation
└── workspace/
```

## Example Schemas

### Input Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "action": {
      "type": "string",
      "enum": ["create", "update", "delete"]
    },
    "target": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9-]*$"
    },
    "options": {
      "type": "object"
    }
  },
  "required": ["action", "target"]
}
```

### Output Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["success", "failure", "partial"]
    },
    "result": {
      "type": "object"
    },
    "errors": {
      "type": "array",
      "items": { "type": "string" }
    }
  },
  "required": ["status"]
}
```

## Configuration

```json
// ayo.json
{
  "squad": {
    "io": {
      "input_schema": "schemas/input.json",
      "output_schema": "schemas/output.json",
      "strict": true  // Fail on validation error
    }
  }
}
```

## Implementation

```go
// internal/squads/validation.go
import "github.com/santhosh-tekuri/jsonschema/v5"

type SchemaValidator struct {
    inputSchema  *jsonschema.Schema
    outputSchema *jsonschema.Schema
    strict       bool
}

func (v *SchemaValidator) ValidateInput(input any) error {
    if v.inputSchema == nil {
        return nil
    }
    
    if err := v.inputSchema.Validate(input); err != nil {
        if v.strict {
            return fmt.Errorf("input validation failed: %w", err)
        }
        log.Warn().Err(err).Msg("input validation warning")
    }
    return nil
}
```

## Error Messages

Clear, actionable errors:

```
Error: Squad input validation failed

  Field: target
  Error: Does not match pattern "^[a-z][a-z0-9-]*$"
  Value: "My-Project"
  
  Hint: Use lowercase letters, numbers, and hyphens only.
        Example: "my-project"
```

## Files to Create/Modify

1. **`internal/squads/validation.go`** (new) - Schema loading and validation
2. **`internal/squads/dispatch.go`** - Call validators
3. **`go.mod`** - Add jsonschema library
4. **`docs/squads.md`** - Document schema usage

## Acceptance Criteria

- [ ] Input schema loaded from schemas/input.json
- [ ] Output schema loaded from schemas/output.json
- [ ] Input validated before dispatch processing
- [ ] Output validated before returning to user
- [ ] Clear error messages with field locations
- [ ] strict: false logs warnings instead of failing
- [ ] Missing schema files are not an error (optional)

## Testing

- Test valid input passes
- Test invalid input fails with clear error
- Test output validation
- Test strict vs non-strict mode
- Test missing schema files (should succeed)
