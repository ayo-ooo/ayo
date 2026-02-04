# Sandbox Debugging Guide

This guide covers troubleshooting and debugging sandbox execution issues in ayo.

## Quick Diagnostics

```bash
# Check daemon status
ayo status

# Start daemon in foreground for debugging
ayo daemon start --foreground

# Enable debug logging
AYO_DEBUG=1 ayo @agent "command"
```

## Common Issues

### Daemon Not Running

**Symptoms:**
- "Connection refused" errors
- Sandbox commands hang

**Solution:**
```bash
# Start the daemon manually
ayo daemon start

# Or start in foreground to see logs
ayo daemon start --foreground
```

### Sandbox Not Acquiring

**Symptoms:**
- "No sandboxes available" error
- "Pool exhausted" message

**Diagnosis:**
```bash
ayo status
# Check "Active sandboxes" count vs "Pool capacity"
```

**Solutions:**
1. Increase pool size in config:
```json
{
  "providers": {
    "sandbox": {
      "pool": {
        "max_size": 8
      }
    }
  }
}
```

2. Release orphaned sandboxes:
```bash
ayo daemon stop
ayo daemon start
```

### Command Execution Failures

**Symptoms:**
- Commands fail with permission errors
- "No such file or directory" inside sandbox

**Diagnosis:**
```bash
# Check mount configuration
ayo agents show @agent-name

# Verify host path exists
ls -la /path/to/mount
```

**Solutions:**
1. Verify mounts in agent config:
```json
{
  "sandbox": {
    "mounts": [
      {"source": ".", "target": "/workspace"}
    ]
  }
}
```

2. Check read-only settings:
```json
{
  "sandbox": {
    "mounts": [
      {"source": ".", "target": "/workspace", "readonly": false}
    ]
  }
}
```

### Network Issues

**Symptoms:**
- `curl` fails inside sandbox
- Package installation fails

**Diagnosis:**
```bash
# Test network from host
curl -I https://example.com

# Check agent network config
ayo agents show @agent-name | grep network
```

**Solutions:**
1. Enable network in agent config:
```json
{
  "sandbox": {
    "network": true
  }
}
```

2. Check container runtime network:
```bash
docker network ls
docker network inspect bridge
```

### Container Not Starting

**Symptoms:**
- "Failed to create sandbox" error
- Container runtime errors

**Diagnosis:**
```bash
# Check Docker is running
docker ps

# Check Lima (macOS)
limactl list

# View container logs
docker logs ayo-sandbox-*
```

**Solutions:**
1. Restart container runtime:
```bash
# Docker
docker restart

# Lima
limactl stop default && limactl start default
```

2. Check image availability:
```bash
docker images | grep busybox
```

## Debug Logging

### Enable Debug Mode

```bash
# Via environment variable
export AYO_DEBUG=1
ayo @agent "command"

# Via CLI flag
ayo --debug @agent "command"
```

### Debug Output Locations

Debug logs appear on stderr and include:
- Daemon IPC messages
- Sandbox lifecycle events
- Command execution details
- Provider operations

### Example Debug Session

```bash
# Start daemon with debug
AYO_DEBUG=1 ayo daemon start --foreground &

# In another terminal, run agent
AYO_DEBUG=1 ayo @agent "ls /workspace"

# Watch for:
# [daemon] Received sandbox.acquire request
# [sandbox] Creating container with image: busybox
# [sandbox] Mounting . -> /workspace
# [sandbox] Executing: ls /workspace
```

## Provider-Specific Issues

### Docker Provider

**Check Docker socket:**
```bash
ls -la /var/run/docker.sock
docker info
```

**Clean up orphaned containers:**
```bash
docker ps -a | grep ayo-sandbox | awk '{print $1}' | xargs docker rm -f
```

### Lima Provider (macOS)

**Check Lima status:**
```bash
limactl list
limactl shell default uname -a
```

**Restart Lima:**
```bash
limactl stop default
limactl start default
```

### None Provider

The `none` provider executes commands directly on the host (no isolation).

**When to use:**
- Development/testing
- CI environments without containers
- Quick debugging

**Configure:**
```json
{
  "providers": {
    "sandbox": {
      "type": "none"
    }
  }
}
```

## Configuration Reference

### Agent Sandbox Config

```json
{
  "sandbox": {
    "enabled": true,
    "provider": "docker",
    "image": "golang:1.22",
    "languages": ["go", "python"],
    "network": true,
    "mounts": [
      {"source": ".", "target": "/workspace", "readonly": false}
    ],
    "resources": {
      "cpus": 2,
      "memory_mb": 2048,
      "disk_mb": 10240,
      "timeout": 300
    }
  }
}
```

### Global Pool Config

```json
{
  "providers": {
    "sandbox": {
      "type": "docker",
      "pool": {
        "min_size": 1,
        "max_size": 4
      }
    }
  }
}
```

## Getting Help

If issues persist:

1. Check the daemon logs: `ayo daemon start --foreground`
2. Enable debug mode: `AYO_DEBUG=1`
3. Verify provider is working: `docker ps` or `limactl list`
4. Check agent config: `ayo agents show @agent-name`
