---
id: am-wil0
status: closed
deps: [am-tzub]
links: []
created: 2026-02-18T03:23:47Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Create ayo-todos plugin struct and registration

Create the basic plugin struct for ayo-todos and register it with the registry.

## Context
- First step of extracting ayo-todos plugin
- Just the skeleton, not the full implementation

## Implementation
```go
// internal/planners/builtin/todos/plugin.go

package todos

import "github.com/anthropic/ayo/internal/planners"

type Plugin struct {
    stateDir string
    state    *State
}

func New() planners.PlannerFactory {
    return func(ctx planners.PlannerContext) (planners.PlannerPlugin, error) {
        return &Plugin{stateDir: ctx.StateDir}, nil
    }
}

func (p *Plugin) Name() string { return "ayo-todos" }
func (p *Plugin) Type() planners.PlannerType { return planners.NearTerm }

func init() {
    planners.DefaultRegistry.Register("ayo-todos", New())
}
```

## Files to Create
- internal/planners/builtin/todos/plugin.go

## Dependencies
- am-tzub (PlannerRegistry)

## Acceptance
- Plugin struct created
- Registered with DefaultRegistry
- Compiles without errors

