# Ayo Sandbox Base Image Specification

## Overview

The ayo sandbox uses a minimal busybox-based container image for agent execution.
This provides a lightweight, security-focused environment with essential POSIX tools.

## Base Image

**Image**: `busybox:stable`

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

## Extended Image: `ayo-sandbox:latest`

For agents requiring more tools, ayo can build an extended image:

```dockerfile
FROM busybox:stable

# Add curl for HTTP operations
COPY --from=curlimages/curl:latest /usr/bin/curl /usr/bin/curl

# Add jq for JSON processing  
COPY --from=ghcr.io/jqlang/jq:latest /jq /usr/bin/jq

# Create ayo user
RUN adduser -D -u 1000 ayo
USER ayo
WORKDIR /workspace
```

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
      "name": "docker",
      "config": {
        "pool": {
          "min_size": 1,
          "max_size": 4
        }
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

## Image Management

### Pulling Images

The installer pulls required images:

```bash
# Base image (always pulled)
docker pull busybox:stable

# Extended image (if sandbox enabled)
docker pull ayo-sandbox:latest
```

### Building Extended Image

```bash
# Build locally
docker build -t ayo-sandbox:latest -f internal/sandbox/Dockerfile .

# Or pull from registry
docker pull ghcr.io/alexcabrera/ayo-sandbox:latest
```

## Implementation Notes

### Provider Selection

The sandbox provider is selected based on:

1. Config (`ayo.json` → `providers.sandbox.name`)
2. Availability detection (Docker installed? macOS 15+?)
3. Fallback to `none` provider (host execution)

### Image Caching

Images are cached locally by Docker. First sandbox creation may be slow
if images need to be pulled.

### Container Lifecycle

1. **Create**: `docker run -d --name ayo-sandbox-{id} ...`
2. **Exec**: `docker exec {container} sh -c "{command}"`
3. **Stop**: `docker stop {container}`
4. **Delete**: `docker rm {container}`
