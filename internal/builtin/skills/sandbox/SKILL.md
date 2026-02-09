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

## Directory Structure

The sandbox provides a consistent directory structure for all agents:

| Path | Purpose | Permissions |
|------|---------|-------------|
| `/home/{agent}/` | Your home directory | Private to agent |
| `/shared/` | Shared files between all agents | World-writable (sticky bit) |
| `/workspaces/{session-id}/` | Current session workspace | Session-specific |
| `/mnt/host/` | Mounted host files | Read-only or read-write |
| `/run/ayo/` | Daemon socket for Matrix communication | Read-write |

### Session Workspace Subdirectories

Each session workspace (`/workspaces/{session-id}/`) contains:

| Subdirectory | Purpose |
|--------------|---------|
| `mounted/` | Files mounted from host project |
| `scratch/` | Temporary working files |
| `shared/` | Session-scoped shared files |

### Checking Available Paths

```bash
# Your home directory
echo $HOME
ls -la ~

# Session workspace
echo $WORKSPACE
ls -la $WORKSPACE

# Shared files
ls -la /shared/

# Host mounts
ls -la /mnt/host/
```

## Environment Variables

The sandbox sets these environment variables for each agent:

| Variable | Description | Example |
|----------|-------------|---------|
| `WORKSPACE` | Current session workspace path | `/workspaces/abc123/` |
| `SESSION_ID` | Current session identifier | `abc123` |
| `AGENT` | Your agent handle | `ayo` |
| `HOME` | Your home directory | `/home/ayo` |

### Using Environment Variables

```bash
# Check your identity
echo "I am agent: $AGENT"

# Work in session workspace
cd $WORKSPACE
mkdir -p scratch
touch scratch/temp.txt

# Create output in shared area
echo "result" > $WORKSPACE/shared/output.json
```

## File Sharing Between Agents

Agents can share files with each other using the shared directories.

### Permanent Sharing (across sessions)

For files that should persist and be accessible to all agents:

```bash
# Copy file to global shared directory
cp myfile.txt /shared/

# Other agents can access it
cat /shared/myfile.txt
```

### Session Sharing (within session)

For files specific to the current session:

```bash
# Copy to session shared directory
cp myfile.txt $WORKSPACE/shared/

# Notify other agents via IRC
msg '#general' "File ready at $WORKSPACE/shared/myfile.txt"
```

### Best Practices for File Sharing

1. **Use descriptive names** - Include your agent handle or purpose
2. **Clean up when done** - Remove files no longer needed
3. **Notify via IRC** - Let other agents know when files are ready
4. **Use atomic writes** - Write to temp file, then rename

```bash
# Atomic write pattern
echo '{"status": "complete"}' > /shared/result.tmp
mv /shared/result.tmp /shared/result.json
msg '#general' "Results available at /shared/result.json"
```

## Working with Host Mounts

Mounted host directories appear under `/mnt/host/` or in your session workspace:

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

## Inter-Agent Communication (IRC)

The sandbox includes an ngircd IRC server for agent-to-agent communication.
Your IRC nickname is automatically set to your agent handle (`$AGENT`).

### IRC Helper Scripts

| Command | Usage | Description |
|---------|-------|-------------|
| `msg` | `msg <target> <message>` | Send message to channel or agent |
| `irc-log` | `irc-log [channel] [lines]` | Read channel log history |
| `irc-join` | `irc-join <channel>` | Join an IRC channel |
| `irc-nick` | `irc-nick <nickname>` | Set your IRC nickname |

### Sending Messages

```bash
# Broadcast to all agents on #general
msg '#general' "Task completed successfully"

# Send private message to another agent
msg '@crush' "Can you review my changes?"

# Share a file notification
msg '#general' "Output ready at /shared/results.json"
```

### Reading Messages

```bash
# Read last 20 messages from #general (default)
irc-log general

# Read last 50 messages
irc-log general 50

# Check session channel
irc-log "session-$SESSION_ID" 100
```

### Joining Channels

```bash
# Join a project-specific channel
irc-join project-alpha

# Join with # prefix (both work)
irc-join '#project-beta'
```

### Default Channels

- `#general` - Main channel for all agents (always join this)
- `#session-{id}` - Session-specific channels

### Coordination Patterns

**Request/Response:**
```bash
# Agent A: Request help
msg '#general' "REQUEST: Need code review for /shared/patch.diff"

# Agent B: Acknowledge and respond
msg '#general' "ACK: Reviewing patch.diff now"
# ... review code ...
msg '#general' "DONE: Review complete, comments in /shared/review.txt"
```

**File Handoff:**
```bash
# Producer agent
cp output.json /shared/analysis-results.json
msg '#general' "READY: Analysis results at /shared/analysis-results.json"

# Consumer agent (after reading irc-log)
cat /shared/analysis-results.json
msg '#general' "ACK: Received analysis results"
```

### IRC Server Details

- **Host:** `localhost`
- **Port:** `6667`
- **Protocol:** IRC (RFC 1459)
- **Max Nick Length:** 32 characters
