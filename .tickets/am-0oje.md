---
id: am-0oje
status: closed
deps: [am-rvt0, am-6ye6]
links: []
created: 2026-02-18T03:15:25Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-8epk
---
# Implement input routing in squad dispatch

Route validated input to appropriate agent (designated or squad lead).

## Context
- After input validation, route to either:
  1. Designated agent (from input_accepts)
  2. Squad lead (@ayo-in-squad)

## Implementation
```go
// internal/squads/dispatch.go

func (s *Squad) routeInput(ctx context.Context, input DispatchInput) error {
    targetAgent := s.Constitution.InputAccepts
    if targetAgent == "" {
        targetAgent = "@ayo" // Squad lead
    }
    
    // Create ticket or direct invoke based on mode
    // ...
}
```

## Files to Modify
- internal/squads/dispatch.go

## Dependencies
- am-rvt0 (dispatch infrastructure)
- am-6ye6 (input_accepts parsing)

## Acceptance
- Input routed to designated agent if specified
- Input routed to squad lead otherwise
- Proper context passed to receiving agent

