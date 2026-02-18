---
id: am-rozh
status: closed
deps: [am-0011]
links: []
created: 2026-02-18T03:24:06Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Add ayo-todos instructions for system prompt

Add Instructions() method to ayo-todos plugin.

## Context
- Instructions tell the agent how to use the todos tool
- Injected into system prompt

## Implementation
```go
// internal/planners/builtin/todos/plugin.go

func (p *Plugin) Instructions() string {
    return `## Near-Term Planning

Use the todos tool to track your immediate work:
- Create todos for multi-step tasks
- Mark one todo as in_progress at a time
- Complete todos as you finish them
- Keep the list focused on current session work
`
}
```

## Files to Modify
- internal/planners/builtin/todos/plugin.go

## Dependencies
- am-0011 (tools implementation)

## Acceptance
- Instructions() returns useful guidance
- Clear, concise text

