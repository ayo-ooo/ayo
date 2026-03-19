# Hooks

Execute scripts at agent lifecycle events.

## Overview

Hooks are executable scripts that run at specific points during agent execution. Use them for logging, notifications, metrics, or side effects.

## Directory Structure

```
my-agent/
├── hooks/
│   ├── agent-start
│   ├── agent-finish
│   ├── agent-error
│   ├── step-start
│   └── step-finish
```

## Hook Types

| Hook | When It Runs |
|------|--------------|
| `agent-start` | Agent begins execution |
| `agent-finish` | Agent completes successfully |
| `agent-error` | Agent encounters an error |
| `step-start` | A skill step begins |
| `step-finish` | A skill step completes |

## Hook Input

Hooks receive JSON on stdin with event data:

### agent-start

```json
{
  "event": "agent-start",
  "timestamp": "2024-01-15T10:30:00Z",
  "agent_name": "my-agent",
  "version": "1.0.0"
}
```

### agent-finish

```json
{
  "event": "agent-finish",
  "timestamp": "2024-01-15T10:30:05Z",
  "output": { ... }
}
```

### agent-error

```json
{
  "event": "agent-error",
  "timestamp": "2024-01-15T10:30:03Z",
  "error": "Error message"
}
```

### step-start

```json
{
  "event": "step-start",
  "timestamp": "2024-01-15T10:30:01Z",
  "step": "analyze"
}
```

### step-finish

```json
{
  "event": "step-finish",
  "timestamp": "2024-01-15T10:30:02Z",
  "step": "analyze",
  "duration_ms": 1234
}
```

## Creating Hooks

Hooks are executable scripts (shell, Python, etc.):

### Example: Logging Hook

```bash
#!/bin/bash
# hooks/agent-start
PAYLOAD=$(cat)
TIMESTAMP=$(echo "$PAYLOAD" | jq -r '.timestamp')
AGENT=$(echo "$PAYLOAD" | jq -r '.agent_name')
echo "[$TIMESTAMP] Agent $AGENT started" >> /var/log/agent.log
```

### Example: Webhook Hook

```bash
#!/bin/bash
# hooks/agent-finish
PAYLOAD=$(cat)
curl -X POST "https://hooks.example.com/complete" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD"
```

### Example: Notification Hook

```bash
#!/bin/bash
# hooks/agent-error
PAYLOAD=$(cat)
ERROR=$(echo "$PAYLOAD" | jq -r '.error')
# Send alert
notify-send "Agent Error" "$ERROR"
```

## Hook Execution

- Hooks run asynchronously (don't block the agent)
- Hook failures are logged but don't fail the agent
- Hooks have access to environment variables
- Hooks run in the order they appear in the directory

## Complete Example: Notifier

From the notifier example:

**hooks/agent-start**:
```bash
#!/bin/bash
PAYLOAD=$(cat)
echo "Agent starting..." >> /tmp/agent.log
```

**hooks/agent-finish**:
```bash
#!/bin/bash
PAYLOAD=$(cat)
TIMESTAMP=$(echo "$PAYLOAD" | jq -r '.timestamp')
echo "[$TIMESTAMP] Agent completed" >> /tmp/agent.log
```

**hooks/agent-error**:
```bash
#!/bin/bash
PAYLOAD=$(cat)
ERROR=$(echo "$PAYLOAD" | jq -r '.error')
echo "ERROR: $ERROR" >> /tmp/agent.log
exit 1
```

## Requirements

- Hooks must be executable (`chmod +x hooks/*`)
- Hooks should handle missing fields gracefully
- Hooks should complete quickly (long operations should be backgrounded)

## Environment Variables

Hooks inherit the environment:

```bash
#!/bin/bash
# Access environment variables
API_KEY=${NOTIFICATION_API_KEY:-""}
LOG_LEVEL=${LOG_LEVEL:-"info"}
```

## Best Practices

1. **Fail gracefully**: Log errors but don't crash
2. **Be fast**: Hooks shouldn't slow down the agent
3. **Handle all fields**: Payload fields may be missing
4. **Use jq for JSON**: Easier than raw parsing
5. **Log to stderr**: Don't pollute stdout

## Security

- Hooks run with the same permissions as the agent
- Validate and sanitize any external inputs
- Don't expose sensitive data in logs

## Next Steps

- [Skills](skills.md) - Add modular capabilities
- [Examples](../examples/notifier.md) - See the notifier example
