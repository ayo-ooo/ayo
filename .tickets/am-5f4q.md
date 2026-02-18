---
id: am-5f4q
status: closed
deps: []
links: []
created: 2026-02-18T03:15:51Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-d01h
---
# Create @ayo-in-squad agent variant

Create a squad-scoped variant of @ayo that acts as squad lead.

## Context
- @ayo-in-squad is a context-specific instantiation of @ayo
- It inherits the squad constitution into its system prompt
- It cannot access resources outside the squad sandbox

## Implementation
```go
// internal/agent/squad_lead.go (new file)

// CreateSquadLead creates an @ayo variant scoped to a specific squad
func CreateSquadLead(baseAyo Agent, constitution *squads.Constitution) (Agent, error) {
    lead := baseAyo // Copy
    lead.Handle = "@ayo"  // Still appears as @ayo
    lead.Config.Sandbox.Enabled = true
    lead.IsSquadLead = true
    lead.SquadName = constitution.SquadName
    
    // Inject constitution into system prompt
    lead.CombinedSystem = squads.InjectConstitution(lead.CombinedSystem, constitution)
    
    // Restrict capabilities
    lead.Config.Delegates = nil // Cannot delegate outside squad
    
    return lead, nil
}
```

## Files to Create
- internal/agent/squad_lead.go

## Files to Modify
- internal/agent/agent.go (add IsSquadLead, SquadName fields)

## Acceptance
- Squad lead created from base @ayo
- Constitution injected into system prompt
- Cannot delegate outside squad

