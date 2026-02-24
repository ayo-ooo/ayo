# Sandbox Configuration Guide

Complete reference for sandbox architecture and configuration.

## Sandbox Providers

### Apple Container (macOS 26+)

Native macOS virtualization using Hypervisor.framework.

**Requirements**:
- macOS 26 or later
- Apple Silicon (ARM64)
- Hypervisor.framework entitlement

**Features**:
- VirtioFS for file mounts
- Native performance
- Linux containers on macOS

### systemd-nspawn (Linux)

Lightweight container using Linux namespaces.

**Requirements**:
- systemd 239 or later
- systemd-container package

**Installation**:
```bash
# Debian/Ubuntu
sudo apt install systemd-container

# Fedora/RHEL
sudo dnf install systemd-container

# Arch
sudo pacman -S systemd
```

**Features**:
- Namespace isolation
- Native Linux performance
- Minimal overhead

## File System Layout

### Inside Sandbox

```
/
├── home/
│   └── {agent}/              # Agent's home directory
├── mnt/
│   └── {username}/           # Read-only host mount
├── output/                   # Safe write zone (syncs to host)
├── shared/                   # Shared between agents (mode 1777)
├── workspaces/               # Project workspaces
│   └── {squad}/              # Squad-specific workspace
├── run/
│   └── ayo/                  # Runtime files
│       └── ayod.sock         # In-sandbox daemon socket
└── tmp/                      # Temporary files
```

### On Host

```
~/.local/share/ayo/
├── sandboxes/
│   ├── ayo/                  # @ayo's dedicated sandbox
│   │   ├── rootfs/           # Root filesystem
│   │   └── home/             # Persistent home
│   └── squads/
│       └── {squad-name}/     # Squad sandboxes
│           ├── rootfs/
│           ├── home/
│           └── workspace/
├── output/                   # Synced from sandbox /output/
├── images/                   # Base images
└── cache/                    # Build cache
```

## Sandbox Configuration

### Global Config

In `~/.config/ayo/config.json`:

```json
{
  "sandbox": {
    "provider": "applecontainer",
    "default_image": "alpine:latest",
    "network": false,
    "resources": {
      "memory": "2G",
      "cpu": "2"
    }
  }
}
```

### Agent Config

In agent `config.json`:

```json
{
  "sandbox": {
    "enabled": true,
    "user": "agent",
    "persist_home": true,
    "image": "alpine:latest",
    "network": false,
    "resources": {
      "memory": "1G",
      "cpu": "1"
    },
    "mounts": [
      "/data:/data:ro"
    ]
  }
}
```

### Squad Config

In squad `ayo.json`:

```json
{
  "sandbox": {
    "image": "ubuntu:22.04",
    "network": true,
    "resources": {
      "memory": "8G",
      "cpu": "4"
    }
  }
}
```

## Configuration Options

### Provider

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `provider` | string | Auto-detect | `applecontainer` or `nspawn` |

### Image

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `image` | string | `alpine:latest` | Base image |

**Available images**:
- `alpine:latest` - Minimal Alpine Linux
- `ubuntu:22.04` - Ubuntu LTS
- Custom images via `ayo sandbox image build`

### Resources

```json
{
  "resources": {
    "memory": "2G",
    "cpu": "2",
    "disk": "10G"
  }
}
```

| Field | Format | Example |
|-------|--------|---------|
| `memory` | Size string | `512M`, `2G`, `4G` |
| `cpu` | Count | `1`, `2`, `4` |
| `disk` | Size string | `5G`, `10G`, `50G` |

### Network

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `network` | bool | `false` | Allow network access |

**Security note**: Network access enables internet connectivity. Use only when needed.

### Mounts

```json
{
  "mounts": [
    "/host/path:/container/path:ro",
    "~/data:/data:rw"
  ]
}
```

**Format**: `source:destination:mode`

| Mode | Description |
|------|-------------|
| `ro` | Read-only |
| `rw` | Read-write |

**Default mounts**:
- `/mnt/{username}` → Home directory (read-only)
- `/output` → Output sync zone (read-write)

### User

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `user` | string | Agent name | Unix user in sandbox |
| `persist_home` | bool | `true` | Persist home directory |

## ayod (In-Sandbox Daemon)

### Purpose

ayod runs inside the sandbox and handles:
- Command execution
- File operations
- Tool responses

### Communication

JSON-RPC over Unix socket at `/run/ayo/ayod.sock`.

### Methods

| Method | Description |
|--------|-------------|
| `exec` | Execute command |
| `read_file` | Read file contents |
| `write_file` | Write file |
| `list_dir` | List directory |
| `stat` | File information |

## Sandbox Lifecycle

### Creation

```bash
# @ayo sandbox created automatically
ayo setup

# Squad sandbox created on first use
ayo squad start my-squad
```

### Starting/Stopping

```bash
# Start sandbox
ayo sandbox start @ayo
ayo squad start my-squad

# Stop sandbox
ayo sandbox stop @ayo
ayo squad stop my-squad
```

### Destruction

```bash
# Destroy squad sandbox
ayo squad destroy my-squad

# Reset @ayo sandbox
ayo sandbox reset @ayo
```

## Agent Users

### User Creation

Each agent gets a Unix user:

```bash
# Sanitized from handle
@backend → user "backend"
@my-agent → user "my_agent"
```

### Home Directory

```
/home/{agent}/
├── .bashrc
├── .profile
└── workspace/ → /workspaces/{squad}/
```

### Permissions

- Home: Owned by agent user
- Workspace: Shared (mode 775)
- Output: World-writable (mode 777)

## Shell Access

### Interactive Shell

```bash
# @ayo sandbox
ayo sandbox shell @ayo

# Squad sandbox
ayo squad shell my-squad

# As specific agent
ayo squad shell my-squad @backend
```

### Execute Command

```bash
# Single command
ayo sandbox exec @ayo "ls -la"

# In squad
ayo squad exec my-squad "cat .tickets/*.md"
```

## Image Management

### List Images

```bash
ayo sandbox image list
```

### Build Custom Image

Create `Dockerfile`:

```dockerfile
FROM alpine:latest
RUN apk add --no-cache go nodejs npm python3
```

Build:
```bash
ayo sandbox image build --name dev-image --file Dockerfile
```

Use in config:
```json
{
  "sandbox": {
    "image": "dev-image"
  }
}
```

## Troubleshooting

### Sandbox creation fails

```bash
# Check provider availability
ayo doctor | grep sandbox

# View detailed error
AYO_DEBUG=1 ayo sandbox start @ayo
```

### Apple Container issues

- Requires macOS 26+
- Requires Apple Silicon
- Check System Preferences → Security

### nspawn issues

```bash
# Check systemd-nspawn
which systemd-nspawn

# Check capabilities
systemd-nspawn --help | head
```

### Mount issues

```bash
# Check source path exists
ls -la /path/to/mount

# Check permissions
stat /path/to/mount
```

### Network not working

```bash
# Verify network enabled
ayo agent show @name | grep network

# Test inside sandbox
ayo sandbox shell @ayo
ping google.com
```

### Disk full

```bash
# Check sandbox disk usage
du -sh ~/.local/share/ayo/sandboxes/

# Clean up
ayo sandbox prune
```

### Performance issues

Reduce resource limits:
```json
{
  "resources": {
    "memory": "1G",
    "cpu": "1"
  }
}
```

Or increase for demanding workloads:
```json
{
  "resources": {
    "memory": "8G",
    "cpu": "4"
  }
}
```
