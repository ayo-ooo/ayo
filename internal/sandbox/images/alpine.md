# Ayo Sandbox Base Image Specification

## Overview

The ayo sandbox uses native container solutions for agent execution.
On macOS 26+, this means Apple Container (github.com/apple/container).
On Linux, native container solutions like systemd-nspawn are supported.

This provides a lightweight, security-focused environment with essential POSIX tools
and a real package manager for installing additional software.

## Base Image

**Image**: `alpine:3.21` (from Docker Hub or compatible OCI registry)

Alpine Linux provides a minimal Linux environment (~5MB) with:
- Real package manager (apk) for installing tools
- musl libc for small footprint
- Full POSIX compliance
- Security-focused design

## Included Tools (Alpine Base)

Alpine includes essential utilities. Key ones used by agents:

### Shell & Scripting
- `sh` - BusyBox ash (POSIX-compatible)
- `ash` - Almquist shell
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

### Network (when networking enabled)
- `wget` - HTTP client
- `ping` - Connectivity testing
- `nc` (netcat) - TCP/UDP connections

### System
- `ps`, `kill` - Process management
- `date` - Date/time
- `sleep` - Delays
- `id`, `whoami` - User info
- `adduser`, `addgroup` - User management

### Package Management
- `apk` - Alpine package manager
  - `apk add <package>` - Install package
  - `apk del <package>` - Remove package
  - `apk search <term>` - Search packages
  - `apk update` - Update package index

## Language Support

Agents can specify required languages in their config:

```json
{
  "sandbox": {
    "languages": ["go", "python", "node"]
  }
}
```

These are installed on first use via the daemon.

### Language Installation

| Language | Install Command | Approximate Size |
|----------|-----------------|------------------|
| Go | `apk add go` | ~200MB |
| Python | `apk add python3 py3-pip` | ~50MB |
| Node.js | `apk add nodejs npm` | ~80MB |
| Ruby | `apk add ruby` | ~40MB |
| Rust | `apk add rust cargo` | ~300MB |

## Configuration

### Agent Config (`config.json`)

```json
{
  "sandbox": {
    "enabled": true,
    "user": "ayo",
    "persist_home": true,
    "network": true,
    "resources": {
      "cpus": 1,
      "memory_mb": 512,
      "disk_mb": 1024,
      "timeout_seconds": 300
    }
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

## Directory Structure

```
/home/{agent}/       # Agent home directories (persistent)
/shared/             # Cross-agent permanent storage
/workspaces/{id}/    # Session-scoped workspaces
/mnt/host/           # Mounted host files
```

## Security Considerations

1. **Agent users**: Each agent runs as dedicated non-root user
2. **Resource limits**: CPU, memory, disk quotas enforced
3. **Network isolation**: Can be disabled per-agent
4. **Read-only mounts**: Sensitive directories mounted read-only
5. **Persistent homes**: Agent state preserved across sessions

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
container run -d --name ayo-sandbox-{id} -v /project:/workspace alpine:3.21 sleep infinity

# Execute command
container exec ayo-sandbox-{id} sh -c "{command}"

# Stop container
container stop ayo-sandbox-{id}

# Delete container
container delete ayo-sandbox-{id}
```

### systemd-nspawn (Linux)

On Linux systems with systemd, nspawn provides lightweight container isolation.

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

1. **Create**: `container run -d --name ayo-sandbox alpine:3.21 sleep infinity`
2. **Exec**: `container exec ayo-sandbox sh -c "{command}"`
3. **Stop**: `container stop ayo-sandbox`
4. **Delete**: `container delete ayo-sandbox`
