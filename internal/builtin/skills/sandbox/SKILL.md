---
name: sandbox
description: Understanding sandbox execution environment. Use when working in sandboxed agents or when execution context matters for command behavior.
compatibility: Requires sandbox-enabled agent configuration
metadata:
  author: ayo
  version: "1.0"
---

# Sandbox Awareness Skill

Understand and work effectively within sandbox execution environments.

## When to Use

Activate this skill when:
- Commands behave differently than expected (missing tools, permissions)
- Working on tasks that may be affected by isolation
- User asks about the execution environment
- Need to understand file system mount points
- Network access issues arise

## Understanding Sandbox Execution

### What Is Sandboxed

When sandbox execution is enabled:
- **Commands run in containers** - Not directly on the host system
- **File system is isolated** - Only mounted directories are accessible
- **Network may be restricted** - Depends on agent configuration
- **Resources are limited** - CPU, memory, disk have quotas

### What Is NOT Sandboxed

- The LLM inference (runs on host/daemon)
- Session and memory storage
- Configuration files

## Detecting Sandbox Context

Check if running in sandbox:

```bash
# Check for container indicators
cat /etc/os-release 2>/dev/null || echo "Not in container"

# Check for sandbox mounts
mount | grep virtiofs
```

## Working with Mounts

Sandbox containers have specific mount points:

| Mount | Purpose | Access |
|-------|---------|--------|
| `/workspace` | Current project directory | Read-write |
| `/data` | Shared data directory | Varies |
| `/tmp` | Temporary files | Read-write |

### Checking Available Mounts

```bash
# List mounted directories
df -h

# Check specific paths
ls -la /workspace
ls -la /data 2>/dev/null
```

## Handling Missing Tools

Sandboxes use minimal base images. Common tools may be missing.

### Available by Default

- `sh`, `bash` (shell)
- `ls`, `cat`, `echo`, `grep`, `sed`, `awk`
- `curl` (for network operations)
- Basic POSIX utilities

### Installing Additional Tools

For Go projects:
```bash
go install github.com/tool/pkg@latest
```

For Python:
```bash
pip install package-name
```

For Node.js:
```bash
npm install -g package-name
```

## Network Considerations

### When Network Is Enabled

- Can fetch dependencies
- Can access external APIs
- Can clone git repositories

### When Network Is Disabled

- Use pre-installed tools only
- Work with files already mounted
- Cannot access external resources

Check network status:
```bash
curl -s -o /dev/null -w "%{http_code}" https://example.com || echo "Network unavailable"
```

## Resource Limits

Be aware of resource constraints:

```bash
# Check available disk space
df -h /workspace

# Check memory (if /proc is mounted)
cat /proc/meminfo 2>/dev/null | head -5
```

## Best Practices

1. **Check environment first** - Understand what's available before starting
2. **Use absolute paths** - Reference mounted directories explicitly
3. **Handle missing tools gracefully** - Check before using specialized tools
4. **Clean up temp files** - Resources are limited
5. **Report environment issues** - If sandbox limits prevent task completion

## Troubleshooting

### Command Not Found

```bash
# Check if tool exists
which {tool} || echo "Not installed"

# Try alternatives
command -v git || command -v svn
```

### Permission Denied

```bash
# Check file permissions
ls -la {path}

# Check if path is in mounted directory
pwd
mount | grep $(pwd)
```

### Network Timeouts

```bash
# Check connectivity
ping -c 1 8.8.8.8 2>&1 || echo "No network"

# Check DNS
nslookup google.com 2>&1 || echo "DNS unavailable"
```
