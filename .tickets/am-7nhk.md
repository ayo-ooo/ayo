---
id: am-7nhk
status: open
deps: [am-mw6n]
links: []
created: 2026-02-18T03:17:35Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-4i9o
---
# Integrate UnifiedSearcher into @ayo dispatch logic

Use UnifiedSearcher when @ayo decides where to route work.

## Context
- @ayo evaluates: trivial? squad match? agent match? do it myself?
- Use embedding similarity to find best match
- Threshold below which @ayo handles it

## Implementation
```go
// internal/run/dispatch.go (new file)

type DispatchDecision struct {
    Target     string  // "@ayo", "@agent", "#squad"
    Confidence float64
    Reason     string
}

func (r *Runner) decideDispatch(ctx context.Context, prompt string) (*DispatchDecision, error) {
    // Check if trivial (heuristic: short prompt, no domain keywords)
    if isTrivial(prompt) {
        return &DispatchDecision{Target: "@ayo", Reason: "trivial task"}, nil
    }
    
    result, err := r.searcher.FindBest(ctx, prompt)
    if err != nil || result == nil || result.Score < 0.5 {
        return &DispatchDecision{Target: "@ayo", Reason: "no good match"}, nil
    }
    
    return &DispatchDecision{
        Target:     result.Handle,
        Confidence: result.Score,
        Reason:     "semantic match",
    }, nil
}
```

## Files to Create
- internal/run/dispatch.go

## Dependencies
- am-mw6n (UnifiedSearcher)

## Acceptance
- @ayo uses embeddings for routing
- Falls back to self if no good match
- Trivial tasks handled directly

