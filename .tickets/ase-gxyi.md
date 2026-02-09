---
id: ase-gxyi
status: closed
deps: [ase-w2n6]
links: []
created: 2026-02-09T03:05:39Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-qnjh
---
# Expose Matrix socket to sandboxes

Mount the Matrix Unix socket into sandboxes so agents can communicate via `ayo chat` CLI.

## Background

Conduit listens on a Unix socket at /tmp/ayo/matrix.sock. Agents in sandboxes need access to communicate with other agents. The daemon socket is already mounted; Matrix socket needs the same treatment.

## Implementation

1. Add Matrix socket to sandbox mount configuration
2. Ensure socket path is consistent across host and sandbox
3. Test that `ayo chat` commands work from inside sandbox

## Mount configuration

In sandbox creation, add:
```go
{
    Source:      '/tmp/ayo/matrix.sock',
    Destination: '/tmp/ayo/matrix.sock',
    Mode:        providers.MountModeBind,
}
```

## Files to modify

- internal/server/sandbox_manager.go (add mount)
- internal/sandbox/apple.go (if platform-specific handling needed)

## Testing

1. Start daemon with Conduit
2. Create sandbox
3. Exec into sandbox: `ayo sandbox exec -- ayo chat rooms`
4. Verify it connects and lists rooms

## Acceptance Criteria

- Matrix socket mounted in sandboxes
- ayo chat commands work from sandbox
- Socket permissions correct

