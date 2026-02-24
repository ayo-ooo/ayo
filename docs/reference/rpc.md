# Daemon RPC Reference

Complete reference for the ayo daemon's JSON-RPC 2.0 API.

## Connection

### Socket Location

| Platform | Path |
|----------|------|
| Unix/macOS | `~/.local/share/ayo/daemon.sock` |
| Windows | `\\.\pipe\ayo-daemon` |

### Protocol

- **Format**: JSON-RPC 2.0
- **Transport**: Unix domain socket (or named pipe on Windows)
- **Framing**: Newline-delimited JSON (each message ends with `\n`)

### Example Request

```json
{"jsonrpc": "2.0", "id": 1, "method": "daemon.ping", "params": {}}
```

### Example Response

```json
{"jsonrpc": "2.0", "id": 1, "result": {"pong": true, "timestamp": "2024-01-15T10:30:00Z"}}
```

## Authentication

The daemon uses Unix socket permissions for access control. No token or credential-based authentication is required. Access is controlled by file system permissions on the socket file.

## Error Codes

### Standard JSON-RPC Errors

| Code | Name | Description |
|------|------|-------------|
| -32700 | Parse Error | Invalid JSON |
| -32600 | Invalid Request | Not a valid JSON-RPC request |
| -32601 | Method Not Found | Unknown method name |
| -32602 | Invalid Params | Invalid method parameters |
| -32603 | Internal Error | Internal server error |

### Application Errors

| Code | Name | Description |
|------|------|-------------|
| -1001 | Sandbox Not Found | Requested sandbox does not exist |
| -1002 | Sandbox Exhausted | Sandbox pool has no available instances |
| -1003 | Sandbox Timeout | Sandbox operation timed out |
| -1004 | Daemon Shutting Down | Daemon is in shutdown state |

## Methods

### Daemon Methods

#### `daemon.ping`

Health check for daemon connectivity.

**Request**: None

**Response**:
```json
{
  "pong": true,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### `daemon.status`

Get daemon status and statistics.

**Request**: None

**Response**:
```json
{
  "version": "0.3.0",
  "uptime": "2h15m30s",
  "active_sandboxes": 3,
  "active_sessions": 2,
  "active_triggers": 5
}
```

#### `daemon.shutdown`

Gracefully shut down the daemon.

**Request**:
```json
{
  "force": false,
  "timeout": 30
}
```

**Response**: Empty object `{}`

### Sandbox Methods

#### `sandbox.acquire`

Acquire a sandbox from the pool.

**Request**:
```json
{
  "provider": "apple",
  "config": {},
  "timeout": 60
}
```

**Response**:
```json
{
  "sandbox_id": "sbx-abc123",
  "provider": "apple",
  "state": "running"
}
```

#### `sandbox.release`

Release a sandbox back to the pool.

**Request**:
```json
{
  "sandbox_id": "sbx-abc123"
}
```

**Response**: Empty object `{}`

#### `sandbox.exec`

Execute a command in a sandbox.

**Request**:
```json
{
  "sandbox_id": "sbx-abc123",
  "command": ["ls", "-la"],
  "working_dir": "/workspace",
  "env": {"DEBUG": "1"},
  "timeout": 30
}
```

**Response**:
```json
{
  "exit_code": 0,
  "stdout": "total 24\ndrwxr-xr-x ...",
  "stderr": "",
  "duration_ms": 45
}
```

#### `sandbox.status`

Get status of all sandboxes.

**Request**: None

**Response**:
```json
{
  "sandboxes": [
    {
      "id": "sbx-abc123",
      "provider": "apple",
      "state": "running",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

#### `sandbox.join`

Join an existing sandbox session.

**Request**:
```json
{
  "sandbox_id": "sbx-abc123"
}
```

**Response**: Empty object `{}`

#### `sandbox.agents`

List agents in a sandbox.

**Request**:
```json
{
  "sandbox_id": "sbx-abc123"
}
```

**Response**:
```json
{
  "agents": ["code", "reviewer"]
}
```

### Session Methods

#### `session.list`

List active sessions.

**Request**: None

**Response**:
```json
{
  "sessions": [
    {
      "id": "sess-xyz789",
      "agent": "code",
      "sandbox_id": "sbx-abc123",
      "started_at": "2024-01-15T10:05:00Z"
    }
  ]
}
```

#### `session.start`

Start a new agent session.

**Request**:
```json
{
  "agent": "code",
  "sandbox_id": "sbx-abc123",
  "config": {}
}
```

**Response**:
```json
{
  "session_id": "sess-xyz789"
}
```

#### `session.stop`

Stop a session.

**Request**:
```json
{
  "session_id": "sess-xyz789"
}
```

**Response**: Empty object `{}`

### Agent Methods

#### `agent.wake`

Wake an agent (activate for work).

**Request**:
```json
{
  "agent": "code",
  "sandbox_id": "sbx-abc123",
  "context": {}
}
```

**Response**:
```json
{
  "status": "awake",
  "session_id": "sess-xyz789"
}
```

#### `agent.sleep`

Put an agent to sleep (deactivate).

**Request**:
```json
{
  "agent": "code",
  "sandbox_id": "sbx-abc123"
}
```

**Response**: Empty object `{}`

#### `agent.status`

Get agent status.

**Request**:
```json
{
  "agent": "code",
  "sandbox_id": "sbx-abc123"
}
```

**Response**:
```json
{
  "agent": "code",
  "state": "awake",
  "current_task": "ayo-123"
}
```

#### `agent.invoke`

Invoke an agent with a prompt.

**Request**:
```json
{
  "agent": "code",
  "sandbox_id": "sbx-abc123",
  "prompt": "Fix the bug in main.go",
  "context": {}
}
```

**Response**:
```json
{
  "response": "I've fixed the null pointer...",
  "tool_calls": 5,
  "tokens_used": 1250
}
```

### Trigger Methods

#### `trigger.list`

List all registered triggers.

**Request**: None

**Response**:
```json
{
  "triggers": [
    {
      "id": "trg-abc123",
      "name": "daily-review",
      "category": "poll",
      "enabled": true,
      "last_fired": "2024-01-15T09:00:00Z"
    }
  ]
}
```

#### `trigger.get`

Get trigger details.

**Request**:
```json
{
  "trigger_id": "trg-abc123"
}
```

**Response**:
```json
{
  "id": "trg-abc123",
  "name": "daily-review",
  "category": "poll",
  "config": {},
  "enabled": true
}
```

#### `trigger.register`

Register a new trigger.

**Request**:
```json
{
  "name": "code-review",
  "category": "watch",
  "agent": "reviewer",
  "config": {
    "path": "/workspace",
    "pattern": "*.go"
  }
}
```

**Response**:
```json
{
  "trigger_id": "trg-def456"
}
```

#### `trigger.remove`

Remove a trigger.

**Request**:
```json
{
  "trigger_id": "trg-abc123"
}
```

**Response**: Empty object `{}`

#### `trigger.test`

Test-fire a trigger.

**Request**:
```json
{
  "trigger_id": "trg-abc123"
}
```

**Response**: Empty object `{}`

#### `trigger.set_enabled`

Enable or disable a trigger.

**Request**:
```json
{
  "trigger_id": "trg-abc123",
  "enabled": false
}
```

**Response**: Empty object `{}`

#### `trigger.history`

Get trigger execution history.

**Request**:
```json
{
  "trigger_id": "trg-abc123",
  "limit": 10
}
```

**Response**:
```json
{
  "executions": [
    {
      "id": "exec-001",
      "fired_at": "2024-01-15T09:00:00Z",
      "status": "success",
      "duration_ms": 5200
    }
  ]
}
```

### Flow Methods

#### `flow.run`

Execute a flow.

**Request**:
```json
{
  "flow": "deploy",
  "inputs": {
    "environment": "staging"
  }
}
```

**Response**:
```json
{
  "run_id": "run-abc123",
  "status": "completed",
  "outputs": {}
}
```

#### `flow.list`

List available flows.

**Request**:
```json
{
  "directory": "/workspace"
}
```

**Response**:
```json
{
  "flows": ["deploy", "test", "review"]
}
```

#### `flow.get`

Get flow definition.

**Request**:
```json
{
  "flow": "deploy"
}
```

**Response**:
```json
{
  "name": "deploy",
  "steps": [],
  "inputs": {}
}
```

#### `flow.history`

Get flow execution history.

**Request**:
```json
{
  "flow": "deploy",
  "limit": 10
}
```

**Response**:
```json
{
  "runs": []
}
```

### Ticket Methods

#### `tickets.create`

Create a new ticket.

**Request**:
```json
{
  "title": "Implement login",
  "description": "Add OAuth support",
  "tags": ["feature"],
  "directory": "/workspace"
}
```

**Response**:
```json
{
  "ticket_id": "ayo-abc1"
}
```

#### `tickets.get`

Get ticket details.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace"
}
```

**Response**:
```json
{
  "id": "ayo-abc1",
  "title": "Implement login",
  "status": "open",
  "assignee": "",
  "tags": ["feature"]
}
```

#### `tickets.list`

List tickets.

**Request**:
```json
{
  "directory": "/workspace",
  "status": "open",
  "assignee": "code"
}
```

**Response**:
```json
{
  "tickets": []
}
```

#### `tickets.update`

Update ticket fields.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace",
  "title": "New title",
  "description": "New description"
}
```

**Response**: Empty object `{}`

#### `tickets.delete`

Delete a ticket.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace"
}
```

**Response**: Empty object `{}`

#### `tickets.start`

Mark ticket as in-progress.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace"
}
```

**Response**: Empty object `{}`

#### `tickets.close`

Close a ticket.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace"
}
```

**Response**: Empty object `{}`

#### `tickets.reopen`

Reopen a closed ticket.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace"
}
```

**Response**: Empty object `{}`

#### `tickets.block`

Mark ticket as blocked.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace"
}
```

**Response**: Empty object `{}`

#### `tickets.assign`

Assign ticket to an agent.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace",
  "assignee": "code"
}
```

**Response**: Empty object `{}`

#### `tickets.add_note`

Add a note to a ticket.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace",
  "note": "Started implementation"
}
```

**Response**: Empty object `{}`

#### `tickets.ready`

Get tickets ready for work.

**Request**:
```json
{
  "directory": "/workspace",
  "agent": "code"
}
```

**Response**:
```json
{
  "tickets": []
}
```

#### `tickets.blocked`

Get blocked tickets.

**Request**:
```json
{
  "directory": "/workspace"
}
```

**Response**:
```json
{
  "tickets": []
}
```

#### `tickets.add_dep`

Add dependency between tickets.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace",
  "depends_on": "ayo-xyz9"
}
```

**Response**: Empty object `{}`

#### `tickets.remove_dep`

Remove dependency between tickets.

**Request**:
```json
{
  "ticket_id": "ayo-abc1",
  "directory": "/workspace",
  "depends_on": "ayo-xyz9"
}
```

**Response**: Empty object `{}`

### Squad Methods

#### `squads.create`

Create a new squad.

**Request**:
```json
{
  "name": "backend-team",
  "description": "Backend development squad",
  "agents": ["code", "reviewer"]
}
```

**Response**:
```json
{
  "squad_id": "squad-abc123",
  "path": "~/.local/share/ayo/sandboxes/squads/backend-team"
}
```

#### `squads.destroy`

Destroy a squad.

**Request**:
```json
{
  "squad_id": "squad-abc123"
}
```

**Response**:
```json
{
  "destroyed": true
}
```

#### `squads.list`

List all squads.

**Request**:
```json
{
  "include_stopped": false
}
```

**Response**:
```json
{
  "squads": []
}
```

#### `squads.get`

Get squad details.

**Request**:
```json
{
  "squad_id": "squad-abc123"
}
```

**Response**:
```json
{
  "id": "squad-abc123",
  "name": "backend-team",
  "state": "running",
  "agents": ["code", "reviewer"]
}
```

#### `squads.start`

Start a squad.

**Request**:
```json
{
  "squad_id": "squad-abc123"
}
```

**Response**:
```json
{
  "started": true
}
```

#### `squads.stop`

Stop a squad.

**Request**:
```json
{
  "squad_id": "squad-abc123"
}
```

**Response**:
```json
{
  "stopped": true
}
```

#### `squads.add_agent`

Add an agent to a squad.

**Request**:
```json
{
  "squad_id": "squad-abc123",
  "agent": "tester"
}
```

**Response**:
```json
{
  "added": true
}
```

#### `squads.remove_agent`

Remove an agent from a squad.

**Request**:
```json
{
  "squad_id": "squad-abc123",
  "agent": "tester"
}
```

**Response**:
```json
{
  "removed": true
}
```

#### `squads.tickets_ready`

Get tickets ready for squad agents.

**Request**:
```json
{
  "squad_id": "squad-abc123"
}
```

**Response**:
```json
{
  "tickets": []
}
```

#### `squads.notify_agents`

Notify agents of new work.

**Request**:
```json
{
  "squad_id": "squad-abc123",
  "message": "New tickets available"
}
```

**Response**:
```json
{
  "notified": ["code", "reviewer"]
}
```

#### `squads.wait_completion`

Wait for squad to complete work.

**Request**:
```json
{
  "squad_id": "squad-abc123",
  "timeout": 3600
}
```

**Response**:
```json
{
  "completed": true,
  "tickets_closed": 5
}
```

#### `squads.sync_output`

Sync squad output to host.

**Request**:
```json
{
  "squad_id": "squad-abc123",
  "destination": "/workspace/output"
}
```

**Response**:
```json
{
  "synced": true,
  "files": 12
}
```

#### `squads.cleanup`

Clean up squad resources.

**Request**:
```json
{
  "squad_id": "squad-abc123"
}
```

**Response**:
```json
{
  "cleaned": true
}
```

#### `squads.dispatch`

Dispatch work to squad agents.

**Request**:
```json
{
  "squad_id": "squad-abc123"
}
```

**Response**:
```json
{
  "dispatched": true,
  "assigned_tickets": 3
}
```

## Client Libraries

### Go Client

```go
import "github.com/anthropics/ayo/internal/daemon"

client, err := daemon.NewClient()
if err != nil {
    log.Fatal(err)
}
defer client.Close()

result, err := client.Call("daemon.ping", nil)
```

### CLI Usage

```bash
# The ayo CLI uses the daemon automatically when available
ayo daemon start
ayo run code "Fix the bug"  # Uses daemon for sandbox management
```

## See Also

- [CLI Reference](cli.md) - Command-line interface
- [Guides: Triggers](../guides/triggers.md) - Setting up triggers
- [Tutorials: Squads](../tutorials/squads.md) - Multi-agent coordination
