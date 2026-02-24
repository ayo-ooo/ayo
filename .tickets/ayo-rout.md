---
id: ayo-rout
status: closed
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-xfu3
tags: [routing, ayo]
---
# Implement @ayo smart routing

Make @ayo intelligent about when to handle tasks directly vs delegate to agents vs dispatch to squads.

## Context

Current state:
- `Dispatcher` does semantic routing via embeddings
- `find_agent` tool exists for explicit discovery
- `UnifiedSearcher` indexes agents/squads
- Static delegation config maps task types to agents

## Routing Decision Flow

```
User: "Build the auth feature"
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│ @AYO ROUTING DECISION                                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Is this trivial? (<100 chars, greeting, clarification)     │
│     └─▶ Handle directly                                        │
│                                                                 │
│  2. Does user explicitly target? (@agent or #squad)            │
│     └─▶ Route to target                                        │
│                                                                 │
│  3. Is there a matching squad? (semantic search, >0.7)         │
│     └─▶ Dispatch to squad (multi-agent coordination)          │
│                                                                 │
│  4. Is there a specialist agent? (semantic search, >0.6)       │
│     └─▶ Invoke agent directly                                  │
│                                                                 │
│  5. Can @ayo handle it with current tools?                     │
│     └─▶ Handle directly                                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Key Principle

> If a task needs multiple agents collaborating, use a squad.
> If a single agent can do it, invoke that agent directly.
> Squads are for coordination, not for single-agent tasks.

## Routing Signals

| Signal | Weight | Notes |
|--------|--------|-------|
| Explicit targeting | Override | User said `@agent` or `#squad` |
| Squad mission match | 0.7+ | Task matches squad SQUAD.md mission |
| Agent description match | 0.6+ | Task matches agent description |
| Task complexity | Heuristic | Multi-file, multi-concern → squad |
| Memory hints | Bonus | Previous similar tasks routed somewhere |

## Implementation

### Router Interface

```go
// internal/routing/router.go
type Router struct {
    searcher *capabilities.UnifiedSearcher
    memory   *memory.Service
}

type RoutingDecision struct {
    Target     string          // "@agent", "#squad", or ""
    TargetType TargetType      // Agent, Squad, Self
    Confidence float64
    Reason     string
}

func (r *Router) Decide(ctx context.Context, input string) (*RoutingDecision, error) {
    // 1. Check for trivial input
    if isTrivial(input) {
        return &RoutingDecision{TargetType: Self, Reason: "trivial input"}, nil
    }
    
    // 2. Check for explicit targeting
    if target := parseExplicitTarget(input); target != "" {
        return &RoutingDecision{Target: target, ...}, nil
    }
    
    // 3. Search for matching squad
    squads, _ := r.searcher.SearchSquadsOnly(ctx, input, 1)
    if len(squads) > 0 && squads[0].Similarity > 0.7 {
        return &RoutingDecision{
            Target:     squads[0].Handle,
            TargetType: Squad,
            Confidence: squads[0].Similarity,
            Reason:     "matched squad mission",
        }, nil
    }
    
    // 4. Search for matching agent
    agents, _ := r.searcher.SearchAgentsOnly(ctx, input, 1)
    if len(agents) > 0 && agents[0].Similarity > 0.6 {
        return &RoutingDecision{
            Target:     agents[0].Handle,
            TargetType: Agent,
            Confidence: agents[0].Similarity,
            Reason:     "matched agent capability",
        }, nil
    }
    
    // 5. Handle directly
    return &RoutingDecision{TargetType: Self, Reason: "no better match"}, nil
}
```

### Integration

Hook into `cmd/ayo/root.go` before agent invocation:

```go
if handle == "@ayo" && !hasExplicitTarget(args) {
    decision, _ := router.Decide(ctx, prompt)
    if decision.TargetType != routing.Self {
        // Redirect to decision.Target
    }
}
```

## Files to Create/Modify

- Create `internal/routing/router.go`
- Create `internal/routing/signals.go` (complexity heuristics)
- Modify `cmd/ayo/root.go` to use router
- Modify `internal/run/dispatch.go` to integrate

## Testing

- Test trivial input handled directly
- Test explicit targeting respected
- Test squad matching above threshold
- Test agent matching above threshold
- Test fallback to @ayo
- Test memory hints influence routing
