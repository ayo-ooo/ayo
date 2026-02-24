---
id: ayo-ao4q
status: open
deps: [ayo-kkxg]
links: []
created: 2026-02-23T23:13:09Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [sandbox, agents]
---
# Implement shared sandbox with per-agent Unix users

Migrate existing per-agent user functionality to ayod and establish the @ayo sandbox as the default shared workspace.

## Context: Existing Functionality

The codebase **already has** per-agent Unix users via `EnsureAgentUser()` in `apple.go:668-723`:
- Sanitizes agent handle → Unix username
- Creates user with `adduser -D -s /bin/sh username`
- Copies dotfiles from host

This ticket is about **migrating** this functionality to ayod, not building from scratch.

## Design

```
@AYO SANDBOX (shared by default):
/home/
├── ayo/              # Unix user: ayo
├── crush/            # Unix user: crush  
├── reviewer/         # Unix user: reviewer
└── {agent-name}/     # Created on first use via ayod

/mnt/{user}/          # Host home (read-only)
/workspace/           # Shared workspace (group writable)
/output/              # Safe write zone
```

## Key Principle: Agents ARE Unix Users

**Previous approach (rejected)**: Agents run as shared `ayo` user with `$HOME` override.

**New approach**: Each agent is a distinct Unix user inside the sandbox:
- `@ayo` → Unix user `ayo`
- `@crush` → Unix user `crush`
- File ownership shows real agent: `ls -la` → `-rw-r--r-- crush crush main.go`

## Benefits

1. **Clear ownership**: `ls -la` shows who created each file
2. **Standard semantics**: No confusion about identity
3. **Process isolation**: Standard Unix mechanisms work
4. **Permissions**: Agents can share via group, restrict via user
5. **Handoff**: Agent can read another's files, ownership is clear

## Implementation via ayod

User creation is handled by ayod (see ayo-kkxg):

```go
// Host daemon requests user creation
client := ayod.Connect(sandboxSocket)
client.UserAdd(ayod.UserAddRequest{
    Username: "crush",
    Shell:    "/bin/sh",
})

// Host daemon executes command as user
client.Exec(ayod.ExecRequest{
    User:    "crush",
    Command: []string{"bash", "-c", script},
    Cwd:     "/workspace",
})
```

## Shared Workspace Permissions

```bash
# /workspace is group-writable so all agents can collaborate
drwxrwxr-x  agents agents  /workspace

# All agent users belong to 'agents' group
$ id crush
uid=1001(crush) gid=1001(crush) groups=1001(crush),100(agents)
```

## Sandbox Lifecycle

1. First invocation creates @ayo sandbox with ayod as PID 1
2. ayod creates `ayo` user automatically
3. When `@crush` is invoked:
   - Host asks ayod to ensure `crush` user exists
   - ayod creates user if needed
   - ayod executes command as `crush`
4. Sandbox persists across sessions

## Files to Modify

- `internal/sandbox/sandbox.go` - Add `GetSharedSandboxID()`, shared vs isolated logic
- `internal/sandbox/bash.go` - Use ayod client for execution
- `internal/sandbox/providers/*.go` - Bootstrap with ayod, remove direct exec

## Testing

- Test user creation via ayod
- Test file ownership shows correct agent
- Test agents can read each other's files
- Test /workspace group permissions
- Test isolated sandbox creates fresh user namespace
