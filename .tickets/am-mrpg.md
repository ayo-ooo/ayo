---
id: am-mrpg
status: closed
deps: []
links: []
created: 2026-02-18T03:14:33Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Add planner plugin type to manifest.json

Extend the plugin system to support planner plugins.

## Context
- Plugin manifest: internal/plugins/manifest.go
- Need to declare planners in plugin manifest

## Implementation
Add to Manifest struct:
```go
type Manifest struct {
    // ... existing fields
    Planners []PlannerDef `json:"planners,omitempty"`
}

type PlannerDef struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"`  // "near" or "long"
    Description string      `json:"description"`
    EntryPoint  string      `json:"entry_point"` // Go plugin .so or binary
}
```

## Files to Modify
- internal/plugins/manifest.go

## Acceptance
- Planners field parsed from manifest.json
- Validation for planner definitions
- Type must be 'near' or 'long'

