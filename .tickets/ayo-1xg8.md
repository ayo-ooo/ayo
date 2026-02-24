---
id: ayo-1xg8
status: open
deps: [ayo-kkxg]
links: []
created: 2026-02-23T22:15:18Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [sandbox, ayo]
---
# Standardize @ayo sandbox bootstrap

Configure the default @ayo sandbox with ayod and proper user setup.

## Bootstrap Sequence

1. Create sandbox with base image (alpine:3.21)
2. Inject ayod binary at `/usr/local/bin/ayod`
3. Start sandbox with ayod as entrypoint (PID 1)
4. ayod creates initial `ayo` user
5. ayod creates shared `agents` group
6. ayod sets up standard directories

## Standard Directories

```
/home/ayo/           # @ayo's home (persistent)
/workspace/          # Shared workspace, group=agents, mode=2775
/output/             # Write zone for host sync
/mnt/                # Mount point for host filesystem
/run/ayod.sock       # ayod communication socket
```

## User Setup

```bash
# @ayo user (created by ayod on bootstrap)
$ id ayo
uid=1000(ayo) gid=1000(ayo) groups=1000(ayo),100(agents)

# agents group for shared workspace access
$ getent group agents
agents:x:100:ayo
```

## Implementation

### Host-side

```go
// internal/sandbox/bootstrap.go
func BootstrapSandbox(provider providers.SandboxProvider, id string) error {
    // 1. Copy ayod binary into sandbox
    ayodPath := getAyodBinaryPath()
    provider.CopyFile(id, ayodPath, "/usr/local/bin/ayod")
    
    // 2. Start sandbox with ayod as entrypoint
    provider.Start(id, "/usr/local/bin/ayod")
    
    // 3. Wait for ayod to be ready
    client := ayod.Connect(socketPath)
    client.WaitReady(timeout)
    
    // 4. Create ayo user
    client.UserAdd("ayo", "/bin/sh")
}
```

### ayod-side

```go
// cmd/ayod/bootstrap.go
func (s *Server) Bootstrap() {
    // Create agents group
    exec.Command("addgroup", "-g", "100", "agents").Run()
    
    // Create standard directories
    os.MkdirAll("/workspace", 0775)
    os.Chown("/workspace", 0, 100)  // root:agents
    
    os.MkdirAll("/output", 0777)
    os.MkdirAll("/mnt", 0755)
}
```

## Persistence

The sandbox is **persistent** (not ephemeral):
- `/home/ayo/` survives daemon restarts
- Installed packages persist
- Agent state accumulates over time

## Files to Modify

- Create `internal/sandbox/bootstrap.go`
- Update `internal/sandbox/providers/apple.go` - Use new bootstrap
- Update `internal/sandbox/providers/linux.go` - Use new bootstrap
- Add ayod binary to distribution/release

## Testing

- Test bootstrap creates all directories
- Test ayo user exists with correct groups
- Test sandbox persistence across daemon restart
- Test ayod socket is accessible
