---
id: ayo-kkxg
status: open
deps: []
links: []
created: 2026-02-24T00:46:06Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ayo-6h19
tags: [sandbox, daemon, ayod]
---
# Implement ayod in-sandbox daemon

Create lightweight Go binary that runs inside sandboxes, providing a clean interface for all sandbox operations.

## Why ayod?

Currently, sandbox operations use ad-hoc `container exec` calls with provider-specific code. This is:
- Fragile (different behavior between Apple Container and systemd-nspawn)
- Limited (can't easily add new capabilities)
- Confusing (agents run as shared user with `$HOME` tricks)

ayod provides a **single entry point** for all sandbox operations with consistent behavior across providers.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  HOST                                                           │
│  ┌─────────────┐     ┌─────────────┐                           │
│  │ ayo CLI     │────▶│ ayo daemon  │                           │
│  └─────────────┘     └──────┬──────┘                           │
│                             │ Unix socket                       │
├─────────────────────────────┼───────────────────────────────────┤
│  SANDBOX                    ▼                                   │
│                      /run/ayod.sock                             │
│                             │                                   │
│                      ┌──────┴──────┐                           │
│                      │   ayod      │  (PID 1)                  │
│                      └──────┬──────┘                           │
│                             │                                   │
│              ┌──────────────┼──────────────┐                   │
│              ▼              ▼              ▼                   │
│         /home/ayo    /home/crush    /home/reviewer             │
│         (user: ayo)  (user: crush)  (user: reviewer)           │
└─────────────────────────────────────────────────────────────────┘
```

## ayod Responsibilities

| Function | RPC Method | Description |
|----------|------------|-------------|
| User management | `UserAdd` | Create agent user with home directory |
| Command execution | `Exec` | Run command as specified user |
| File request proxy | `FileRequest` | Proxy file_request to host daemon |
| Output sync | `SyncOutput` | Copy /output contents to host |
| Health check | `Health` | Report sandbox status |
| Shell access | `Shell` | Interactive shell for debugging |

## Implementation

### Binary Location
- Build: `cmd/ayod/main.go`
- Install: `/usr/local/bin/ayod` inside sandbox
- Size target: <5MB (statically compiled)

### Entry Point

```go
// cmd/ayod/main.go
package main

import (
    "net"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // Create socket
    listener, _ := net.Listen("unix", "/run/ayod.sock")
    
    // Handle signals for clean shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
    
    // Start RPC server
    server := NewServer()
    go server.Serve(listener)
    
    // Wait for shutdown signal
    <-sigCh
    server.Shutdown()
}
```

### RPC Interface

```go
// internal/ayod/rpc.go
type AyodService struct{}

type UserAddRequest struct {
    Username string `json:"username"`
    Shell    string `json:"shell"`
}

type ExecRequest struct {
    User    string            `json:"user"`
    Command []string          `json:"command"`
    Env     map[string]string `json:"env"`
    Cwd     string            `json:"cwd"`
}

type ExecResponse struct {
    ExitCode int    `json:"exit_code"`
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
}

func (s *AyodService) UserAdd(req UserAddRequest) error
func (s *AyodService) Exec(req ExecRequest) (ExecResponse, error)
func (s *AyodService) Health() (HealthResponse, error)
```

### Sandbox Bootstrap

1. Host creates sandbox with base image
2. Host copies `ayod` binary to sandbox at `/usr/local/bin/ayod`
3. Host starts sandbox with `ayod` as entrypoint (not `sleep infinity`)
4. ayod creates `/run/ayod.sock` and listens for connections
5. Host mounts socket for bidirectional communication

### Provider Changes

Update `internal/sandbox/providers/` to use ayod:

```go
// Before (apple.go)
cmd := exec.CommandContext(ctx, "container", "exec", id, "--user", user, ...)

// After
client := ayod.Connect(socketPath)
result, _ := client.Exec(ayod.ExecRequest{
    User:    user,
    Command: command,
    Env:     env,
})
```

## Files to Create

- `cmd/ayod/main.go` - Entry point
- `internal/ayod/server.go` - RPC server
- `internal/ayod/client.go` - RPC client for host
- `internal/ayod/user.go` - User management
- `internal/ayod/exec.go` - Command execution
- `internal/ayod/types.go` - Shared types

## Files to Modify

- `internal/sandbox/providers/apple.go` - Use ayod client
- `internal/sandbox/providers/linux.go` - Use ayod client
- `internal/sandbox/bash.go` - Use ayod for execution
- Build scripts to include ayod in distribution

## Testing

- Unit tests for RPC handlers
- Integration test: create sandbox, run commands via ayod
- Test user creation and isolation
- Test graceful shutdown
- Cross-platform testing (macOS + Linux)

## Future Extensions

Once ayod exists, we can easily add:
- Package installation (`ayod pkg install git`)
- Network configuration
- Resource limits
- GPU access
- Container-to-container networking for squad agents
