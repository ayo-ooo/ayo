---
id: ayo-plex
status: open
deps: []
links: []
created: 2026-02-23T12:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-plug
tags: [plugins, planners]
---
# Task: External Planner Loading Completion

## Summary

Complete the external planner loading mechanism that currently has a TODO placeholder at `internal/plugins/planners.go:100-107`. This will allow plugins to provide custom near-term and long-term planners via Go plugin (.so) files.

## Context

The planner plugin interface is well-defined in `internal/planners/interface.go`:

```go
type PlannerPlugin interface {
    Name() string
    Type() PlannerType  // "near" or "long"
    Init(ctx context.Context) error
    Close() error
    Tools() []fantasy.AgentTool
    Instructions() string
    StateDir() string
}
```

However, `internal/plugins/planners.go` has this TODO:

```go
// TODO: Support loading planners from external plugins
// This would use Go plugin mechanism to load .so files
// that implement the PlannerPlugin interface
```

Currently only builtin planners (`ayo-todos`, `ayo-tickets`) work.

## Technical Approach

### Go Plugin Loading

```go
func LoadExternalPlanner(path string) (planners.PlannerPlugin, error) {
    plug, err := plugin.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open plugin: %w", err)
    }
    
    // Look for NewPlanner symbol
    sym, err := plug.Lookup("NewPlanner")
    if err != nil {
        return nil, fmt.Errorf("lookup NewPlanner: %w", err)
    }
    
    // Assert it's the right type
    newFn, ok := sym.(func() planners.PlannerPlugin)
    if !ok {
        return nil, fmt.Errorf("NewPlanner has wrong signature")
    }
    
    return newFn(), nil
}
```

### Plugin Manifest Entry

```json
{
  "components": {
    "planners": {
      "custom-planner": {
        "path": "bin/planner.so",
        "type": "near",
        "description": "Custom near-term planner"
      }
    }
  }
}
```

### State Directory

External planners get state directories at:
```
~/.local/share/ayo/sandboxes/{sandbox}/.planner.{name}/
```

## Implementation Steps

1. [ ] Implement `LoadExternalPlanner()` function in `internal/plugins/planners.go`
2. [ ] Add planner loading to plugin registry discovery
3. [ ] Update `GetPlannerByType()` to check external planners
4. [ ] Add planner validation (type, interface compliance)
5. [ ] Handle planner state directory setup for externals
6. [ ] Add tests with mock external planner
7. [ ] Document external planner development in docs/planners.md

## Dependencies

- Depends on: None (can be done independently)
- Blocks: `ayo-plrg` (plugin registry improvements)

## Acceptance Criteria

- [ ] External .so plugins can be loaded as planners
- [ ] External planners get proper state directories
- [ ] Error handling for invalid/missing plugins
- [ ] At least one test external planner works
- [ ] Documentation covers external planner development

## Files to Modify

- `internal/plugins/planners.go` - Add loading logic
- `internal/plugins/registry.go` - Register discovered planners
- `internal/planners/manager.go` - Use external planners
- `docs/planners.md` - Document external development

## Notes

- Go plugins have cross-compilation limitations (same Go version, same OS/arch)
- Consider alternative: exec-based plugins (like tools) for cross-platform support
- May want to support both .so and exec-based planners

---

*Created: 2026-02-23*
