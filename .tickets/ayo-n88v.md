---
id: ayo-n88v
status: closed
deps: [ayo-9k8m]
links: []
created: 2026-02-23T22:15:47Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-xfu3
tags: [squads, routing]
---
# Implement squad dispatch routing

When a dispatch is sent to a squad, route it to the appropriate agent based on explicit targeting, input_accepts configuration, or default to lead.

## Context

Squad dispatches need intelligent routing to the right agent. This ticket implements the routing logic.

## Routing Priority

1. **Explicit agent targeting**: `@agent #squad` syntax routes directly
2. **input_accepts**: If configured in ayo.json, route to specified agent
3. **Lead agent**: Default fallback for unrouted dispatches

## Syntax Examples

```bash
# Route to specific agent in squad
ayo dispatch @frontend #dev-team "Build the login page"

# Route to squad (uses input_accepts or lead)
ayo dispatch #dev-team "Implement user auth"

# Short form for active squad
ayo @frontend "Build the login page"
```

## Configuration

```json
// ayo.json for squad
{
  "squad": {
    "lead": "@architect",
    "input_accepts": "@planner",  // Optional: different from lead
    "agents": ["@frontend", "@backend", "@qa"]
  }
}
```

## Implementation

```go
// internal/squads/dispatch.go
func (s *SquadService) RouteDispatch(dispatch *Dispatch) (string, error) {
    // 1. Check for explicit agent targeting
    if dispatch.TargetAgent != "" {
        if !s.hasAgent(dispatch.TargetAgent) {
            return "", fmt.Errorf("agent %s not in squad", dispatch.TargetAgent)
        }
        return dispatch.TargetAgent, nil
    }
    
    // 2. Check for input_accepts
    if s.config.InputAccepts != "" {
        return s.config.InputAccepts, nil
    }
    
    // 3. Default to lead
    if s.config.Lead != "" {
        return s.config.Lead, nil
    }
    
    return "", fmt.Errorf("no routing target for dispatch")
}
```

## Dispatch Flow

```
User: ayo dispatch #dev-team "Build auth"
       │
       ▼
┌─────────────────────┐
│ Parse dispatch      │
│ Squad: dev-team     │
│ Agent: (none)       │
└─────────────────────┘
       │
       ▼
┌─────────────────────┐
│ Check input_accepts │──▶ @planner (if configured)
└─────────────────────┘
       │ (not configured)
       ▼
┌─────────────────────┐
│ Route to lead       │──▶ @architect
└─────────────────────┘
```

## Files to Modify

1. **`internal/squads/dispatch.go`** - Routing logic
2. **`internal/squads/service.go`** - Dispatch handling
3. **`cmd/ayo/dispatch.go`** - Parse @agent #squad syntax
4. **`internal/squads/config.go`** - InputAccepts field

## Acceptance Criteria

- [ ] Explicit @agent #squad routes to agent
- [ ] input_accepts routes non-targeted dispatches
- [ ] Lead receives dispatches when no input_accepts
- [ ] Error when targeting non-existent agent
- [ ] Error when squad has no valid routing target
- [ ] Syntax parsing handles all variations

## Testing

- Test explicit agent routing
- Test input_accepts routing
- Test lead fallback routing
- Test invalid agent error
- Test syntax parsing variations
