---
id: am-kjh7
status: closed
deps: [am-2u05]
links: []
created: 2026-02-18T03:20:15Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Add planner state directory to sandbox mounts

Mount planner state directories into sandbox containers.

## Context
- Planners store state in sandbox directory
- State dirs need to be accessible inside container
- Near-term: .planner.near/, Long-term: .planner.long/

## Implementation
```go
// internal/sandbox/mounts.go

func getPlannerMounts(sandboxDir string) []providers.Mount {
    return []providers.Mount{
        {Source: filepath.Join(sandboxDir, ".planner.near"), Target: "/.planner.near"},
        {Source: filepath.Join(sandboxDir, ".planner.long"), Target: "/.planner.long"},
    }
}
```

## Files to Modify
- internal/sandbox/squad.go (add mounts)
- internal/sandbox/ayo.go (add mounts)

## Dependencies
- am-2u05 (SandboxPlannerManager)

## Acceptance
- Planner state dirs mounted
- State persists across container restarts
- Both near and long-term dirs available

