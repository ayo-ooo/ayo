---
id: am-et0u
status: closed
deps: []
links: []
created: 2026-02-18T03:16:22Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-11v2
---
# Create agent home directories in @ayo sandbox

Set up home directories for agents that run directly under @ayo (not in squads).

## Context
- @ayo sandbox exists at 'ayo-orchestrator'
- Agents invoked directly should have persistent homes
- Location: ~/.local/share/ayo/sandboxes/ayo/home/{agent}/

## Implementation
```go
// internal/sandbox/ayo.go

func EnsureAgentHome(ctx context.Context, provider *AppleProvider, agentHandle string) (string, error) {
    homeDir := paths.AyoAgentHomeDir(agentHandle)
    if err := os.MkdirAll(homeDir, 0755); err != nil {
        return "", err
    }
    return homeDir, nil
}
```

Update paths:
```go
// internal/paths/paths.go

func AyoAgentHomeDir(agentHandle string) string {
    return filepath.Join(AyoSandboxDir(), "home", strings.TrimPrefix(agentHandle, "@"))
}
```

## Files to Modify
- internal/sandbox/ayo.go
- internal/paths/paths.go

## Acceptance
- Home directory created on first agent invocation
- Persists across sessions
- Mounted into sandbox container

