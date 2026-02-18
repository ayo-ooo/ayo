---
id: am-xpxb
status: closed
deps: [am-et0u]
links: []
created: 2026-02-18T03:16:29Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-11v2
---
# Mount agent home into @ayo sandbox

Mount agent home directories when invoking agents in @ayo sandbox.

## Context
- When @ayo invokes an agent directly, mount that agent's home
- Agent sees /home/{agent}/ in container

## Implementation
```go
// internal/run/run.go

func (r *Runner) setupAgentMounts(ag agent.Agent) []providers.Mount {
    mounts := []providers.Mount{} // existing mounts
    
    if !ag.IsSquadLead {
        // Direct invocation under @ayo
        homeDir, _ := sandbox.EnsureAgentHome(ctx, r.sandboxProvider, ag.Handle)
        mounts = append(mounts, providers.Mount{
            Source: homeDir,
            Target: "/home/" + strings.TrimPrefix(ag.Handle, "@"),
        })
    }
    
    return mounts
}
```

## Files to Modify
- internal/run/run.go

## Dependencies
- am-et0u (home directory creation)

## Acceptance
- Agent home mounted on direct invocation
- Agent can read/write to home
- Home persists across sessions

