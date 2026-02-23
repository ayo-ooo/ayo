---
id: am-9eec
status: closed
deps: []
links: []
created: 2026-02-20T02:50:34Z
type: task
priority: 2
assignee: Alex Cabrera
tags: [squads, schemas]
---
# Validate squad input against input.jsonschema before dispatch

Squad I/O schemas (input.jsonschema, output.jsonschema) exist in squad directories but validation is not enforced during dispatch. The ValidateInput function should check against the schema if present.

## Resolution

All acceptance criteria now complete:

1. **Dispatch validation** - Already implemented in `internal/squads/dispatch.go:57-75`
2. **CLI command** - Added `ayo squad schema validate-input <squad> <input.json>` in `cmd/ayo/squad.go`

### Implementation

Added `squadSchemaValidateInputCmd()` function that:
- Loads squad schemas via `squads.LoadSquadSchemas()`
- Parses input JSON file
- Validates against input schema via `schema.ValidateAgainstSchema()`
- Supports `--json` output format
- Returns success/failure with clear error messages

### Usage

```bash
# Validate input data against squad's input.jsonschema
ayo squad schema validate-input dev-team input.json

# JSON output
ayo squad schema validate-input dev-team input.json --json
```

## Acceptance Criteria

- ✅ Dispatch with invalid JSON input returns schema validation error
- ✅ Dispatch with valid input proceeds normally
- ✅ 'ayo squad schema validate-input' command validates test input
