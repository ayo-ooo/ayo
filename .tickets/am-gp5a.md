---
id: am-gp5a
status: open
deps: [am-tzub, am-uw8z]
links: []
created: 2026-02-18T03:24:17Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Create ayo-tickets plugin struct and registration

Create the basic plugin struct for ayo-tickets and register it with the registry.

## Context
- First step of extracting ayo-tickets plugin
- Wraps existing internal/tickets/service.go

## Implementation
```go
// internal/planners/builtin/tickets/plugin.go

package tickets

import "github.com/anthropic/ayo/internal/planners"

type Plugin struct {
    stateDir string
    service  *tickets.Service
}

func New() planners.PlannerFactory {
    return func(ctx planners.PlannerContext) (planners.PlannerPlugin, error) {
        svc, err := tickets.NewService(ctx.StateDir)
        if err != nil {
            return nil, err
        }
        return &Plugin{stateDir: ctx.StateDir, service: svc}, nil
    }
}

func (p *Plugin) Name() string { return "ayo-tickets" }
func (p *Plugin) Type() planners.PlannerType { return planners.LongTerm }

func init() {
    planners.DefaultRegistry.Register("ayo-tickets", New())
}
```

## Files to Create
- internal/planners/builtin/tickets/plugin.go

## Dependencies
- am-tzub (PlannerRegistry)

## Acceptance
- Plugin struct created
- Wraps existing tickets.Service
- Registered with DefaultRegistry

