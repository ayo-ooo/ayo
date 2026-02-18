---
id: am-ek2o
status: closed
deps: []
links: []
created: 2026-02-18T03:18:03Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-hin9
---
# Add squad handle normalization

Create utility functions for normalizing squad handles.

## Context
- Similar to agent.NormalizeHandle for @ prefix
- Handle #frontend-team and frontend-team interchangeably

## Implementation
```go
// internal/squads/handle.go (new file)

const SquadPrefix = "#"

func NormalizeHandle(handle string) string {
    if strings.HasPrefix(handle, SquadPrefix) {
        return handle
    }
    return SquadPrefix + handle
}

func StripPrefix(handle string) string {
    return strings.TrimPrefix(handle, SquadPrefix)
}

func IsSquadHandle(s string) bool {
    return strings.HasPrefix(s, SquadPrefix)
}
```

## Files to Create
- internal/squads/handle.go

## Acceptance
- NormalizeHandle adds # if missing
- StripPrefix removes #
- IsSquadHandle detects squad handles

