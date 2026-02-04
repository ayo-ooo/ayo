# Ayo Daemon Architecture

## Overview

The ayo daemon is a background process that manages long-running resources for the ayo CLI. It provides:

1. **Sandbox Pool Management** - Keeps warm sandboxes ready for agent execution
2. **LLM Connection Pooling** - Maintains connections to LLM providers
3. **Memory Index** - Keeps embedding index warm for fast memory retrieval
4. **Session Continuity** - Enables seamless session resumption

## Design Principles

1. **Transparent** - Users don't need to know the daemon exists
2. **On-Demand** - Auto-starts when needed, auto-stops when idle
3. **Resilient** - Graceful degradation if daemon unavailable
4. **Lightweight** - Minimal resource usage when idle

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         ayo CLI                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  DaemonClient                                             │  │
│  │  - Connect() / Disconnect()                               │  │
│  │  - AcquireSandbox(agent) -> SandboxHandle                │  │
│  │  - ReleaseSandbox(handle)                                │  │
│  │  - GetStatus() -> DaemonStatus                           │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ Unix Socket / Named Pipe
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        ayo daemon                               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Server                                                   │  │
│  │  - Listen on socket                                       │  │
│  │  - Handle client connections                              │  │
│  │  - Route requests to managers                             │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                  │
│         ┌────────────────────┼────────────────────┐            │
│         ▼                    ▼                    ▼            │
│  ┌────────────┐      ┌────────────┐      ┌────────────┐       │
│  │ Sandbox    │      │ LLM        │      │ Memory     │       │
│  │ Supervisor │      │ Pool       │      │ Index      │       │
│  │            │      │            │      │            │       │
│  │ - Pool     │      │ - Conns    │      │ - Embed    │       │
│  │ - Acquire  │      │ - Health   │      │ - Search   │       │
│  │ - Release  │      │ - Reuse    │      │ - Cache    │       │
│  └────────────┘      └────────────┘      └────────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

## IPC Protocol

Uses a simple JSON-RPC 2.0 over Unix socket.

**Socket Location:**
- Unix: `~/.local/share/ayo/daemon.sock`
- Windows: `\\.\pipe\ayo-daemon`

**Request Format:**
```json
{
  "jsonrpc": "2.0",
  "method": "sandbox.acquire",
  "params": {"agent": "@ayo", "timeout": 30},
  "id": 1
}
```

**Response Format:**
```json
{
  "jsonrpc": "2.0",
  "result": {"sandbox_id": "abc123", "working_dir": "/sandbox"},
  "id": 1
}
```

## Methods

### Sandbox Management

| Method | Params | Result |
|--------|--------|--------|
| `sandbox.acquire` | `{agent, timeout}` | `{sandbox_id, working_dir}` |
| `sandbox.release` | `{sandbox_id}` | `{}` |
| `sandbox.exec` | `{sandbox_id, command, timeout}` | `{stdout, stderr, exit_code}` |
| `sandbox.status` | `{}` | `{total, idle, in_use}` |

### Daemon Management

| Method | Params | Result |
|--------|--------|--------|
| `daemon.status` | `{}` | `{uptime, sandboxes, memory_usage}` |
| `daemon.shutdown` | `{graceful}` | `{}` |
| `daemon.ping` | `{}` | `{pong: true}` |

## Lifecycle

### Auto-Start

When CLI needs daemon resources:

1. Check if daemon is running (try to connect)
2. If not running, start daemon in background
3. Wait for daemon to be ready (up to 5s)
4. Connect and proceed

### Auto-Stop

Daemon shuts down after configurable idle period (default: 30 min):

1. No active connections
2. No sandboxes in use
3. No pending operations
4. Graceful shutdown with resource cleanup

### Graceful Degradation

If daemon unavailable:

1. CLI falls back to direct execution (no pooling)
2. Sandboxes created on-demand
3. Warning logged but operation continues

## Configuration

```json
// ~/.config/ayo/ayo.json
{
  "daemon": {
    "enabled": true,
    "auto_start": true,
    "idle_timeout": "30m",
    "log_level": "info",
    "socket_path": ""  // Override default socket path
  }
}
```

## Files

- `~/.local/share/ayo/daemon.sock` - Unix socket
- `~/.local/share/ayo/daemon.pid` - PID file
- `~/.local/share/ayo/daemon.log` - Log file (when not attached to terminal)

## Implementation Plan

1. **Phase 1: Core Infrastructure**
   - `internal/daemon/daemon.go` - Main daemon process
   - `internal/daemon/server.go` - IPC server
   - `internal/daemon/client.go` - Client for CLI
   - `internal/daemon/protocol.go` - JSON-RPC protocol

2. **Phase 2: Sandbox Integration**
   - Connect sandbox pool to daemon
   - Implement sandbox.* methods
   - Add CLI integration

3. **Phase 3: Transparent Management**
   - Auto-start/stop logic
   - PID file management
   - Status command (`ayo status`)

4. **Phase 4: Additional Services** (Future)
   - LLM connection pooling
   - Memory index caching
