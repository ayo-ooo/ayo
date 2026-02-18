---
id: am-0yb4
status: closed
deps: [am-5f4q]
links: []
created: 2026-02-18T03:16:05Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-d01h
---
# Add squad lead spawning to squad startup

Automatically spawn squad lead when squad starts.

## Context
- When a squad is started, the squad lead should be available
- Squad lead is the first agent to receive dispatched work

## Implementation
```go
// internal/squads/service.go

func (s *Service) Start(ctx context.Context, name string) error {
    squad, err := s.Get(name)
    // ... existing startup
    
    // Spawn squad lead
    leadAgent, err := agent.CreateSquadLead(ayoAgent, squad.Constitution)
    if err != nil {
        return fmt.Errorf("failed to create squad lead: %w", err)
    }
    
    squad.Lead = leadAgent
    // ... start lead session
}
```

## Files to Modify
- internal/squads/service.go

## Dependencies
- am-5f4q (squad lead creation)

## Acceptance
- Squad lead spawned on squad start
- Squad lead available for input routing
- Squad lead persists while squad is running

