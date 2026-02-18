---
id: am-plyy
status: closed
deps: [am-d2rd]
links: []
created: 2026-02-18T03:19:20Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-okzf
---
# Add on-demand squad startup in flow runtime

Start squads on-demand when flow steps target them.

## Context
- Flow may target a squad that isn't running
- Flow runtime should start squad as needed
- Squad should be available before step executes

## Implementation
```go
// internal/flows/invoker.go

func (i *SandboxAwareInvoker) ensureSquadRunning(ctx context.Context, squad string) error {
    status, err := i.daemonClient.GetSquadStatus(ctx, squad)
    if err != nil {
        return err
    }
    
    if status.State != "running" {
        if err := i.daemonClient.StartSquad(ctx, squad); err != nil {
            return fmt.Errorf("failed to start squad %s: %w", squad, err)
        }
        
        // Wait for squad to be ready
        return i.waitForSquadReady(ctx, squad)
    }
    
    return nil
}
```

## Files to Modify
- internal/flows/invoker.go

## Dependencies
- am-d2rd (squad context invoker)

## Acceptance
- Squad started if not running
- Waits for squad ready before proceeding
- Error if squad fails to start

