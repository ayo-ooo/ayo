# Ayo Sandbox Base Image Specification

## Overview

The ayo sandbox uses native container solutions for agent execution.
On macOS 26+, this means Apple Container (github.com/apple/container).
On Linux, native container solutions like LXC will be supported.

This provides a lightweight, security-focused environment with essential POSIX tools.

## Base Image

**Image**: `busybox:stable` (from Docker Hub or compatible OCI registry)

BusyBox provides a minimal Unix-like environment (~1.5MB) with essential utilities.
It's used as the default sandbox image for speed and security.

## Included Tools (BusyBox Built-ins)

BusyBox includes ~300+ applets. Key ones used by agents:

### Shell & Scripting
- `sh` - Almquist shell (POSIX-compatible)
- `ash` - Almquist shell alias
- `echo`, `printf` - Output
- `test`, `[` - Conditionals
- `expr` - Expression evaluation
- `xargs` - Build command lines from stdin
- `env` - Environment handling

### File Operations
- `ls`, `cat`, `head`, `tail` - View files
- `cp`, `mv`, `rm`, `mkdir`, `rmdir` - Manage files
- `chmod`, `chown` - Permissions
- `ln` - Links
- `find` - Find files
- `touch` - Create/update timestamps
- `wc`, `sort`, `uniq` - Text processing

### Text Processing
- `grep` - Pattern matching
- `sed` - Stream editing
- `awk` - Text processing
- `cut`, `tr` - Field extraction
- `diff`, `patch` - Diff/patch

### Archive & Compression
- `tar` - Archives
- `gzip`, `gunzip` - Compression
- `unzip` - ZIP extraction

### Network (when networking enabled)
- `wget` - HTTP client
- `ping` - Connectivity testing
- `nc` (netcat) - TCP/UDP connections

### System
- `ps`, `kill` - Process management
- `date` - Date/time
- `sleep` - Delays
- `id`, `whoami` - User info

## Language Support

Agents can specify required languages in their config:

```json
{
  "sandbox": {
    "languages": ["go", "python", "node"]
  }
}
```

These are installed on first use via the daemon, creating agent-specific images.

### Language Images

| Language | Base Image | Install Size |
|----------|-----------|--------------|
| Go | `golang:alpine` | ~300MB |
| Python | `python:alpine` | ~50MB |
| Node.js | `node:alpine` | ~120MB |
| Ruby | `ruby:alpine` | ~80MB |
| Rust | `rust:alpine` | ~700MB |

## Configuration

### Agent Config (`config.json`)

```json
{
  "sandbox": {
    "enabled": true,
    "image": "busybox:stable",
    "languages": [],
    "network": true,
    "resources": {
      "cpus": 1,
      "memory_mb": 512,
      "disk_mb": 1024,
      "timeout_seconds": 300
    },
    "mounts": [
      {"source": ".", "target": "/workspace", "read_only": false}
    ]
  }
}
```

### Pool Config (`ayo.json`)

```json
{
  "providers": {
    "sandbox": {
      "type": "apple-container",
      "pool": {
        "min_size": 1,
        "max_size": 4
      }
    }
  }
}
```

## Security Considerations

1. **No root access**: Containers run as non-root user
2. **Resource limits**: CPU, memory, disk quotas enforced
3. **Network isolation**: Can be disabled per-agent
4. **Read-only mounts**: Sensitive directories mounted read-only
5. **Ephemeral**: Containers are destroyed after use

## Sandbox Providers

### Apple Container (macOS 26+)

Apple Container uses the Virtualization.framework to run Linux containers natively on Apple Silicon.

**Requirements:**
- macOS 26 or later (Tahoe)
- Apple Silicon (M1, M2, M3, or later)
- Container service running: `container system start`

**Commands:**
```bash
# Start a sandbox container
container run -d --name ayo-sandbox-{id} -v /project:/workspace busybox:stable sleep infinity

# Execute command
container exec ayo-sandbox-{id} sh -c "{command}"

# Stop container
container stop ayo-sandbox-{id}

# Delete container
container delete ayo-sandbox-{id}
```

### None Provider (Fallback)

When no container provider is available, commands execute directly on the host.
This provides no isolation but allows the sandbox system to function.

## Provider Selection

The sandbox provider is selected automatically based on:

1. Platform detection (macOS + Apple Silicon → Apple Container)
2. Availability check (Is `container` command available? Is service running?)
3. Fallback to `none` provider (host execution without isolation)

## Image Caching

Images are cached locally by the container runtime. First sandbox creation may be slow
if images need to be pulled.

## Container Lifecycle

### Apple Container

1. **Create**: `container run -d --name ayo-sandbox-{id} -v ... busybox:stable sleep infinity`
2. **Exec**: `container exec ayo-sandbox-{id} sh -c "{command}"`
3. **Stop**: `container stop ayo-sandbox-{id}`
4. **Delete**: `container delete ayo-sandbox-{id}`
