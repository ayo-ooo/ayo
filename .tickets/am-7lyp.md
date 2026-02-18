---
id: am-7lyp
status: closed
deps: [am-2u05]
links: []
created: 2026-02-18T03:16:36Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-11v2
---
# Initialize planners for @ayo sandbox

Set up @ayo's own near-term and long-term planners.

## Context
- @ayo needs its own planners to track delegated work
- Planners stored in @ayo sandbox directory
- Uses SandboxPlannerManager with @ayo sandbox context

## Implementation
```go
// internal/sandbox/ayo.go

func (s *AyoSandbox) InitPlanners(manager *planners.SandboxPlannerManager) error {
    planners, err := manager.GetPlanners("ayo", s.Dir, nil) // Use global defaults
    if err != nil {
        return err
    }
    s.Planners = planners
    return nil
}
```

## Files to Modify
- internal/sandbox/ayo.go

## Dependencies
- am-2u05 (SandboxPlannerManager)

## Acceptance
- @ayo has near-term planner (todos)
- @ayo has long-term planner (tickets)
- Planners persist in @ayo sandbox directory

