---
id: am-q2ni
status: closed
deps: []
links: []
created: 2026-02-18T03:14:56Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-8epk
---
# Add schema loading to squad initialization

Load input.jsonschema and output.jsonschema when initializing a squad.

## Context
- Squad initialization: internal/squads/service.go
- Schemas are optional files in squad directory
- Location: ~/.local/share/ayo/sandboxes/squads/{name}/input.jsonschema

## Implementation
```go
// internal/squads/schema.go (new file)

type SquadSchemas struct {
    Input  *jsonschema.Schema // nil if no input.jsonschema
    Output *jsonschema.Schema // nil if no output.jsonschema
}

func LoadSquadSchemas(squadDir string) (*SquadSchemas, error)
```

Integrate into Squad struct in service.go:
```go
type Squad struct {
    // ... existing fields
    Schemas *SquadSchemas
}
```

## Files to Create
- internal/squads/schema.go

## Files to Modify
- internal/squads/service.go (load schemas on Create/Get)

## Acceptance
- Schemas loaded if files exist
- Nil schemas if files don't exist (free-form mode)
- JSON Schema validation via existing schema package

