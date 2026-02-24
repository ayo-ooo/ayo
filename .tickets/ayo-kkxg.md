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
- Confusing (though agents already run as real Unix users via `EnsureAgentUser()`)

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

## Bootstrap Strategy (No `sleep infinity`)

The key insight is that we avoid `sleep infinity` entirely by using a **staged image build** or **single bootstrap exec**:

### Option 1: Derived Image (Preferred)

Build a derived image at ayo install/init time:

```dockerfile
FROM alpine:3.21
COPY ayod /usr/local/bin/ayod
RUN chmod +x /usr/local/bin/ayod
ENTRYPOINT ["/usr/local/bin/ayod"]
```

- Image stored in `~/.local/share/ayo/images/ayo-base-alpine.tar`
- Container starts with ayod as PID 1 immediately
- No intermediate state

### Option 2: Single Bootstrap Exec (Fallback)

If provider doesn't support custom images (e.g., Apple Container limitations):

```
1. Start container with base image + mounted socket directory
2. Single bootstrap exec: inject ayod binary via host tools
3. ayod starts as PID 1 via the provider's mechanism
4. All subsequent operations go through ayod socket
```

For Apple Container specifically:
```bash
# Copy ayod into container filesystem before starting
container copy ayod ${CONTAINER_ID}:/usr/local/bin/ayod

# Start container with ayod as entrypoint
container run --entrypoint /usr/local/bin/ayod ...
```

### Bootstrap Sequence

```
┌─────────────────────────────────────────────────────────────────┐
│ AYOD BOOTSTRAP                                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. ayo checks for derived image in ~/.local/share/ayo/images/ │
│     ├─ If exists: use it                                       │
│     └─ If not: build it (first run or `ayo init`)             │
│                                                                 │
│  2. Container created from derived image                        │
│     → /usr/local/bin/ayod is already present                   │
│     → Entrypoint is ayod                                        │
│                                                                 │
│  3. Container starts                                            │
│     → ayod runs as PID 1                                        │
│     → ayod creates /run/ayod.sock                              │
│     → ayod waits for connections                               │
│                                                                 │
│  4. Host daemon connects to socket                              │
│     → Socket mounted at predictable path                       │
│     → All operations go through socket                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## ayod Responsibilities (MVP)

Focus on core functionality first:

| Function | RPC Method | Description | Priority |
|----------|------------|-------------|----------|
| User management | `UserAdd` | Create agent user with home directory | **MVP** |
| Command execution | `Exec` | Run command as specified user | **MVP** |
| Health check | `Health` | Report sandbox status | **MVP** |
| File read | `ReadFile` | Read file content (no shell) | MVP |
| File write | `WriteFile` | Write file content (no shell) | MVP |

### Deferred to Future

| Function | RPC Method | Description |
|----------|------------|-------------|
| File request proxy | `FileRequest` | Proxy file_request to host daemon |
| Output sync | `SyncOutput` | Copy /output contents to host |
| Shell access | `Shell` | Interactive shell for debugging |
| Package install | `PkgInstall` | Install packages on demand |

## Implementation

### Binary Location
- Build: `cmd/ayod/main.go`
- Install: `/usr/local/bin/ayod` inside sandbox
- Size target: <5MB (statically compiled)
- Build tags: `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`

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
    // Create socket directory if needed
    os.MkdirAll("/run", 0755)
    
    // Remove stale socket
    os.Remove("/run/ayod.sock")
    
    // Create socket
    listener, err := net.Listen("unix", "/run/ayod.sock")
    if err != nil {
        log.Fatal(err)
    }
    
    // Make socket world-accessible (all sandbox users can connect)
    os.Chmod("/run/ayod.sock", 0666)
    
    // Initialize user manager
    users := NewUserManager()
    
    // Handle signals for clean shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
    
    // Start RPC server
    server := NewServer(users)
    go server.Serve(listener)
    
    // Stay alive as PID 1
    <-sigCh
    server.Shutdown()
}
```

### RPC Interface

```go
// internal/ayod/rpc.go
type AyodService struct {
    users *UserManager
}

type UserAddRequest struct {
    Username string `json:"username"`
    Shell    string `json:"shell"`
    Dotfiles []byte `json:"dotfiles,omitempty"` // Optional tar of dotfiles
}

type ExecRequest struct {
    User    string            `json:"user"`
    Command []string          `json:"command"`
    Env     map[string]string `json:"env"`
    Cwd     string            `json:"cwd"`
    Timeout int               `json:"timeout"` // Seconds, 0 = no timeout
}

type ExecResponse struct {
    ExitCode int    `json:"exit_code"`
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
}

func (s *AyodService) UserAdd(req UserAddRequest) error
func (s *AyodService) Exec(req ExecRequest) (ExecResponse, error)
func (s *AyodService) Health() (HealthResponse, error)
func (s *AyodService) ReadFile(path string) ([]byte, error)
func (s *AyodService) WriteFile(path string, content []byte, mode os.FileMode) error
```

### Provider Changes

Update `internal/sandbox/providers/` to use ayod:

```go
// Before (apple.go)
cmd := exec.CommandContext(ctx, "container", "exec", id, "--user", user, ...)

// After
client := ayod.Connect(socketPath)
result, err := client.Exec(ayod.ExecRequest{
    User:    user,
    Command: command,
    Env:     env,
})
```

## Migration from EnsureAgentUser

Current `EnsureAgentUser()` in `apple.go:668-723` does:
1. Sanitize handle → username
2. `adduser -D -s /bin/sh username`
3. Copy dotfiles from host

ayod's `UserAdd` will do the same, but:
- Called via socket instead of `container exec`
- More efficient (no shell overhead)
- Consistent across providers

## Files to Create

- `cmd/ayod/main.go` - Entry point
- `internal/ayod/server.go` - RPC server
- `internal/ayod/client.go` - RPC client for host
- `internal/ayod/user.go` - User management
- `internal/ayod/exec.go` - Command execution
- `internal/ayod/types.go` - Shared types
- `internal/ayod/file.go` - File operations

## Files to Modify

- `internal/sandbox/providers/apple.go` - Use ayod client
- `internal/sandbox/providers/linux.go` - Use ayod client
- `internal/sandbox/bash.go` - Use ayod for execution
- Build scripts to include ayod in distribution
- `Makefile` - Add ayod build target

## Testing

- Unit tests for RPC handlers
- Integration test: create sandbox, run commands via ayod
- Test user creation and isolation
- Test graceful shutdown
- Test socket permissions
- Cross-platform testing (macOS + Linux)
- Test derived image build

## Future Extensions

Once ayod exists, we can easily add:
- Package installation (`ayod pkg install git`)
- Network configuration
- Resource limits
- GPU access
- Container-to-container networking for squad agents
- Process supervision
- Internal file watching
