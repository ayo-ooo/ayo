---
id: am-yfaq
status: closed
deps: [am-hsum, am-ek2o, am-rvt0, am-vego]
links: []
created: 2026-02-18T03:18:10Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-hin9
---
# Implement synchronous squad invocation

Implement the execution path for ayo #squad 'prompt'.

## Context
- Squad invocation should be synchronous (wait for result)
- Start squad if not running
- Route through squad's dispatch system

## Implementation
```go
// cmd/ayo/root.go or internal/run/squad_invoke.go

func invokeSquad(ctx context.Context, handle, prompt string) error {
    squadName := squads.StripPrefix(handle)
    
    // Ensure squad is running
    squad, err := daemonClient.GetOrStartSquad(ctx, squadName)
    if err != nil {
        return fmt.Errorf("failed to start squad: %w", err)
    }
    
    // Dispatch and wait
    result, err := squad.DispatchSync(ctx, DispatchInput{Prompt: prompt})
    if err != nil {
        return err
    }
    
    // Print result
    fmt.Println(result.Output)
    return nil
}
```

## Files to Modify
- cmd/ayo/root.go

## Dependencies
- am-hsum (CLI parsing)
- am-ek2o (handle normalization)
- am-rvt0 (dispatch infrastructure)

## Acceptance
- ayo #squad 'prompt' starts squad if needed
- Waits for completion
- Prints result to stdout

