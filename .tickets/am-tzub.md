---
id: am-tzub
status: closed
deps: [am-vl6l]
links: []
created: 2026-02-18T03:13:29Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Create PlannerRegistry for plugin management

Create a registry that manages planner plugin registration, lookup, and instantiation.

## Context
- Location: internal/planners/registry.go (new file)
- Registry is a singleton that holds all available planner plugins
- Plugins register themselves (built-in) or are loaded from plugin dirs

## Implementation
```go
type Registry struct {
    mu       sync.RWMutex
    plugins  map[string]PlannerFactory
}

type PlannerFactory func(ctx PlannerContext) (PlannerPlugin, error)

func (r *Registry) Register(name string, factory PlannerFactory)
func (r *Registry) Get(name string) (PlannerFactory, bool)
func (r *Registry) List() []string
func (r *Registry) Instantiate(name string, ctx PlannerContext) (PlannerPlugin, error)

var DefaultRegistry = &Registry{...}
```

## Files to Create
- internal/planners/registry.go

## Dependencies
- am-vl6l (PlannerPlugin interface)

## Acceptance
- Registry can register plugins
- Registry can instantiate plugins with context
- Thread-safe access

