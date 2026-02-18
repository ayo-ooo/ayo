---
id: am-hsum
status: closed
deps: []
links: []
created: 2026-02-18T03:17:58Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-hin9
---
# Parse # symbol for squad invocation in CLI

Modify CLI argument parsing to recognize #squad syntax.

## Context
- Current parsing: @ prefix for agents (cmd/ayo/root.go)
- Add # prefix for squads
- Symmetric syntax: ayo #frontend-team 'prompt'

## Implementation
```go
// cmd/ayo/root.go

func parseInvocation(args []string) (invocationType, handle, prompt string) {
    if len(args) == 0 {
        return "agent", "@ayo", ""
    }
    
    first := args[0]
    if strings.HasPrefix(first, "@") {
        return "agent", first, strings.Join(args[1:], " ")
    }
    if strings.HasPrefix(first, "#") {
        return "squad", first, strings.Join(args[1:], " ")
    }
    
    // Default to @ayo
    return "agent", "@ayo", strings.Join(args, " ")
}
```

## Files to Modify
- cmd/ayo/root.go

## Acceptance
- ayo #squad 'prompt' parsed correctly
- ayo @agent 'prompt' still works
- ayo 'prompt' defaults to @ayo

