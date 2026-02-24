---
id: ayo-dupl
status: open
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-6h19
tags: [cleanup, tech-debt]
---
# Consolidate duplicate interfaces and types

Extract shared interfaces to a common package to eliminate duplication.

## Duplicate Interfaces

### AgentInvoker (3 locations)

Identical interface in:
- `internal/squads/invoker.go:10`
- `internal/flows/yaml_executor.go:18`
- `internal/daemon/invoker.go:16`

**Solution**: Create `internal/interfaces/invoker.go`

```go
package interfaces

type AgentInvoker interface {
    Invoke(ctx context.Context, handle, prompt string) (string, error)
}
```

Update all packages to import from `interfaces`.

### Todo types (2 locations)

Duplicate types:
- `internal/run/todo.go` - `Todo`, `TodoStatus`, `TodoItem`
- `internal/ui/todo.go` - `UITodo` (comment: "mirrors run.Todo to avoid import cycles")

**Solution**: Create `internal/types/todo.go` with shared types

## Steps

1. Create `internal/interfaces/` package
2. Move shared interfaces
3. Update imports in all packages
4. Create `internal/types/` if needed
5. Run tests

## Testing

- All tests pass
- No import cycles
