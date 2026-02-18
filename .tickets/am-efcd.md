---
id: am-efcd
status: open
deps: [am-wil0]
links: []
created: 2026-02-18T03:23:54Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Implement ayo-todos state persistence

Add state persistence for the ayo-todos plugin.

## Context
- Todos need to persist in the sandbox's state directory
- JSON file storage for simplicity

## Implementation
```go
// internal/planners/builtin/todos/state.go

type State struct {
    Todos []Todo `json:"todos"`
}

type Todo struct {
    ID         string `json:"id"`
    Content    string `json:"content"`
    ActiveForm string `json:"active_form"`
    Status     string `json:"status"` // pending, in_progress, completed
}

func (p *Plugin) loadState() error
func (p *Plugin) saveState() error
```

## Files to Create
- internal/planners/builtin/todos/state.go

## Dependencies
- am-wil0 (plugin skeleton)

## Acceptance
- State loads from StateDir on Init
- State saves after modifications
- JSON format used

