# Linux Native Container Integration

This document describes how ayo integrates with native Linux container solutions for sandboxed agent execution.

## Chosen Solution: systemd-nspawn

After evaluating LXC, LXD, and systemd-nspawn, we chose **systemd-nspawn** as the primary Linux solution because:

1. **Built into systemd** - No extra packages needed on most modern Linux distros
2. **Lightweight** - Minimal overhead compared to full container runtimes
3. **Simple CLI** - Straightforward command-line interface
4. **Good security** - Proper namespace and cgroup isolation

## Requirements

- **Linux with systemd** (Ubuntu 16.04+, Fedora 21+, Debian 8+, Arch, RHEL 7+, etc.)
- **Root access** or proper permissions for unprivileged containers
- **Container images** in `/var/lib/machines/` or custom directory

## CLI Command Reference

Based on systemd-nspawn and machinectl.

### Container Lifecycle

```bash
# Run a command in a container (ephemeral, removes after exit)
systemd-nspawn -D /path/to/rootfs /bin/sh -c "echo hello"

# Run as a managed machine (persistent)
systemd-nspawn -M mycontainer -D /path/to/rootfs --boot

# Using machinectl for service-managed containers
machinectl start mycontainer
machinectl stop mycontainer
machinectl kill mycontainer
machinectl remove mycontainer
```

### Executing Commands

```bash
# Execute command in running container
machinectl shell mycontainer /bin/sh -c "echo hello"

# Or with systemd-nspawn directly (ephemeral)
systemd-nspawn -D /path/to/rootfs /bin/sh -c "command"
```

### Container Inspection

```bash
# List containers
machinectl list
machinectl list-images

# Show container status
machinectl status mycontainer

# Show container info
machinectl show mycontainer
```

### Image Management

```bash
# Pull an image (requires systemd-importd)
machinectl pull-tar https://example.com/image.tar.xz myimage
machinectl pull-raw https://example.com/image.raw myimage

# Import local tarball
machinectl import-tar /path/to/rootfs.tar.gz myimage

# Clone/copy image
machinectl clone source-image target-image

# Remove image
machinectl remove myimage
```

## Mount Syntax

systemd-nspawn uses `--bind` for mounts:

```bash
# Read-write bind mount
systemd-nspawn --bind=/host/path:/container/path -D /rootfs /bin/sh

# Read-only bind mount
systemd-nspawn --bind-ro=/host/path:/container/path -D /rootfs /bin/sh

# Tmpfs mount
systemd-nspawn --tmpfs=/tmp -D /rootfs /bin/sh
```

## Resource Limits

systemd-nspawn uses systemd properties for resource limits:

```bash
# CPU limit (via systemd slice)
systemd-nspawn --property=CPUQuota=200% -D /rootfs /bin/sh

# Memory limit
systemd-nspawn --property=MemoryMax=2G -D /rootfs /bin/sh

# Combined
systemd-nspawn --property=CPUQuota=100% --property=MemoryMax=1G -D /rootfs /bin/sh
```

## Network Configuration

```bash
# Private network (isolated)
systemd-nspawn --private-network -D /rootfs /bin/sh

# Host network (shared)
systemd-nspawn --network-host -D /rootfs /bin/sh

# Virtual ethernet (recommended for isolation with connectivity)
systemd-nspawn --network-veth -D /rootfs /bin/sh
```

## Ayo Integration

### Check Availability

```go
func isLinuxContainerAvailable() bool {
    if runtime.GOOS != "linux" {
        return false
    }
    
    // Check if systemd-nspawn exists
    cmd := exec.Command("systemd-nspawn", "--version")
    if err := cmd.Run(); err != nil {
        return false
    }
    
    return true
}
```

### Create Sandbox

For ayo, we use an ephemeral approach with a minimal BusyBox rootfs:

```go
// Create a temporary rootfs directory
// Use systemd-nspawn in ephemeral mode
args := []string{
    "--directory=" + rootfsPath,
    "--machine=" + name,
    "--ephemeral",      // Don't persist changes
    "--quiet",          // Less verbose
}

// Add mounts
for _, m := range opts.Mounts {
    if m.ReadOnly {
        args = append(args, fmt.Sprintf("--bind-ro=%s:%s", m.Source, m.Destination))
    } else {
        args = append(args, fmt.Sprintf("--bind=%s:%s", m.Source, m.Destination))
    }
}

// Network
if !opts.Network.Enabled {
    args = append(args, "--private-network")
}

// Resource limits
if opts.Resources.CPUs > 0 {
    args = append(args, fmt.Sprintf("--property=CPUQuota=%d%%", opts.Resources.CPUs*100))
}
if opts.Resources.MemoryMB > 0 {
    args = append(args, fmt.Sprintf("--property=MemoryMax=%dM", opts.Resources.MemoryMB))
}

// Run with keepalive
args = append(args, "/bin/sh", "-c", "sleep infinity")

cmd := exec.CommandContext(ctx, "systemd-nspawn", args...)
```

### Execute Command

```go
// For ephemeral containers, we run a new systemd-nspawn with the command
args := []string{
    "--directory=" + rootfsPath,
    "--machine=" + name,
    "--quiet",
}

// Add working directory via bind mount if needed
if opts.WorkingDir != "" {
    args = append(args, "--chdir=" + opts.WorkingDir)
}

// Add environment variables
for k, v := range opts.Env {
    args = append(args, fmt.Sprintf("--setenv=%s=%s", k, v))
}

// Add command
args = append(args, "/bin/sh", "-c", opts.Command)

cmd := exec.CommandContext(ctx, "systemd-nspawn", args...)
```

## Rootfs Preparation

For BusyBox-based sandboxes, we need a minimal rootfs:

```bash
# Create minimal rootfs directory
mkdir -p /var/lib/ayo/rootfs/{bin,dev,etc,proc,sys,tmp,workspace}

# Extract busybox static binary
# Download from https://busybox.net/downloads/binaries/
cp busybox /var/lib/ayo/rootfs/bin/
cd /var/lib/ayo/rootfs/bin
for cmd in $(./busybox --list); do ln -s busybox $cmd; done

# Create minimal /etc
echo "root:x:0:0:root:/root:/bin/sh" > /var/lib/ayo/rootfs/etc/passwd
echo "root:x:0:" > /var/lib/ayo/rootfs/etc/group
```

The ayo installer will prepare this rootfs automatically.

## Fallback Behavior

If systemd-nspawn is not available:
1. Check for unprivileged user namespaces (for rootless operation)
2. Fall back to NoneProvider (host execution without isolation)

## Distribution Notes

| Distribution | Package | Notes |
|--------------|---------|-------|
| Ubuntu/Debian | `systemd-container` | May need to install separately |
| Fedora/RHEL | Built-in | Included with systemd |
| Arch | Built-in | Included with systemd |
| Alpine | N/A | No systemd, falls back to None |

## Testing

```bash
# Test availability
systemd-nspawn --version

# Test basic functionality (needs root or sudo)
sudo systemd-nspawn -D /path/to/busybox-rootfs /bin/sh -c "echo hello"

# Test with mount
sudo systemd-nspawn --bind=/tmp:/mnt -D /path/to/rootfs /bin/ls /mnt
```

## Notes

- systemd-nspawn typically requires root for full functionality
- Unprivileged containers require user namespace support in the kernel
- The ephemeral mode (`--ephemeral`) is useful for ayo's use case
- Container cleanup is automatic with ephemeral mode
