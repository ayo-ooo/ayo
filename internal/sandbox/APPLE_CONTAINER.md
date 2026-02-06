# Apple Container Integration

This document describes how ayo integrates with Apple Container for sandboxed agent execution on macOS.

## Requirements

- **macOS 26 or later** (Tahoe)
- **Apple Silicon** (M1, M2, M3, or later)
- **Apple Container installed** from [GitHub releases](https://github.com/apple/container/releases)
- **Container service running**: `container system start`

## CLI Command Reference

Based on Apple Container v0.4.x. The CLI is `container` (not `docker`).

### System Management

```bash
# Start the container service (required before first use)
container system start

# Check if service is running
container system status

# Stop the service
container system stop

# Show version information
container system version
```

### Container Lifecycle

```bash
# Run a container (foreground)
container run <image> [command]

# Run a container (detached/background)
container run -d --name <name> <image>

# Create container without starting
container create --name <name> <image>

# Start a created container
container start <container-id>

# Stop a running container
container stop <container-id>
container stop --time 10 <container-id>  # 10 second timeout

# Force stop
container kill <container-id>

# Delete a container
container delete <container-id>
container delete --force <container-id>  # Force delete running container
# Aliases: container rm, container remove
```

### Executing Commands

```bash
# Execute command in running container
container exec <container-id> <command> [args...]

# With working directory
container exec --workdir /app <container-id> <command>

# With environment variables
container exec --env KEY=VALUE <container-id> <command>

# Interactive with TTY
container exec -it <container-id> /bin/sh
```

### Container Inspection

```bash
# List containers (running only)
container list

# List all containers (including stopped)
container list --all

# Get container details (JSON)
container inspect <container-id>

# Get container logs
container logs <container-id>
container logs --follow <container-id>
```

### Image Management

```bash
# Pull an image
container image pull <image>

# List images
container image list

# Delete an image
container image delete <image>
```

## Mount Syntax

Apple Container uses `-v` or `--volume` for bind mounts:

```bash
# Basic mount
container run -v /host/path:/container/path <image>

# Read-only mount
container run -v /host/path:/container/path:ro <image>

# Using --mount (more explicit)
container run --mount type=bind,source=/host/path,target=/container/path <image>
container run --mount type=bind,source=/host/path,target=/container/path,readonly <image>
```

## Resource Limits

```bash
# CPU limit
container run --cpus 2 <image>

# Memory limit (supports K, M, G, T, P suffixes)
container run --memory 2G <image>

# Combined
container run --cpus 2 --memory 2G <image>
```

## Network Configuration

```bash
# Default network (enabled)
container run <image>

# Disable networking
container run --no-dns <image>

# Attach to specific network
container run --network <network-name> <image>
```

## Differences from Current Implementation

The current `internal/sandbox/apple.go` implementation has these issues:

1. **Incorrect container creation**: Uses separate `create` + `start` commands, which is fine, but the command syntax may differ.

2. **Status checking**: Uses `container inspect` and parses for status strings. Need to verify JSON output format.

3. **Service detection**: Uses `container system status` which is correct.

## Ayo Integration Points

### Create Sandbox

```go
// Create and start a container
args := []string{"run", "-d", "--name", name}

// Add mounts
for _, m := range opts.Mounts {
    mountArg := fmt.Sprintf("%s:%s", m.Source, m.Destination)
    if m.ReadOnly {
        mountArg += ":ro"
    }
    args = append(args, "-v", mountArg)
}

// Resource limits
if opts.Resources.CPUs > 0 {
    args = append(args, "--cpus", fmt.Sprintf("%d", opts.Resources.CPUs))
}
if opts.Resources.MemoryMB > 0 {
    args = append(args, "--memory", fmt.Sprintf("%dM", opts.Resources.MemoryMB))
}

// Network (disabled if not enabled)
if !opts.Network.Enabled {
    args = append(args, "--no-dns")
}

// Image and keepalive command
args = append(args, image, "sh", "-c", "sleep infinity")

cmd := exec.CommandContext(ctx, "container", args...)
```

### Execute Command

```go
args := []string{"exec"}

if opts.WorkingDir != "" {
    args = append(args, "--workdir", opts.WorkingDir)
}

for k, v := range opts.Env {
    args = append(args, "--env", fmt.Sprintf("%s=%s", k, v))
}

args = append(args, containerID, "sh", "-c", opts.Command)

cmd := exec.CommandContext(ctx, "container", args...)
```

### Check Availability

```go
func isAppleContainerAvailable() bool {
    // Check platform
    if runtime.GOOS != "darwin" {
        return false
    }
    if runtime.GOARCH != "arm64" {
        return false
    }
    
    // Check if container command exists and service is running
    cmd := exec.Command("container", "system", "status")
    if err := cmd.Run(); err != nil {
        return false
    }
    
    return true
}
```

## Testing

```bash
# Test availability
container system status

# Test basic functionality
container run --rm busybox:stable echo "hello"

# Test with mount
container run --rm -v /tmp:/mnt busybox:stable ls /mnt

# Test exec
container run -d --name test-exec busybox:stable sleep 3600
container exec test-exec echo "hello from exec"
container stop test-exec
container delete test-exec
```

## Notes

- Apple Container runs Linux containers via Virtualization.framework
- Uses virtiofs for fast file sharing with the host
- Images are OCI-compatible, can pull from Docker Hub and other registries
- Container IDs can use the `--name` value (like Docker)
- The `container` CLI requires the service to be started first
