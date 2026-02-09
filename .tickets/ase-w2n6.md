---
id: ase-w2n6
status: open
deps: []
links: []
created: 2026-02-09T03:04:49Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-qnjh
---
# Integrate Conduit Matrix homeserver

Embed Conduit (lightweight Rust Matrix homeserver) as a subprocess managed by the ayo service.

## Background

Matrix is used for inter-agent communication because:
- Auditable message history
- Human can join and observe/participate
- Async message delivery
- Room-based organization (one room per session)

Conduit chosen because:
- Single binary (~15MB)
- SQLite backend (no PostgreSQL)
- Minimal resource usage
- Can be distributed with ayo

## Implementation

1. Download/bundle Conduit binary for target platforms
2. Add Conduit process management to daemon:
   - Start on service start
   - Configure to use Unix socket (not TCP) for local-only access
   - Store data in ~/.local/share/ayo/matrix/
   - Health check and restart on crash
3. Generate server signing keys on first run
4. Configure homeserver name as 'ayo.local'

## Conduit configuration

```toml
[global]
server_name = 'ayo.local'
database_backend = 'sqlite'
database_path = '~/.local/share/ayo/matrix/conduit.db'
port = 0  # Unix socket instead
unix_socket_path = '/tmp/ayo/matrix.sock'
allow_registration = true  # Agents self-register
```

## Files to modify/create

- internal/daemon/conduit.go (new - process management)
- internal/daemon/server.go (start Conduit on daemon start)
- build/conduit/ (new - binary download scripts)
- internal/paths/paths.go (add MatrixDataDir, MatrixSocket)

## Platform considerations

- macOS: Download from Conduit releases
- Linux: Same
- Consider bundling in release builds

## Acceptance Criteria

- Conduit starts with daemon
- Uses Unix socket, not TCP
- Data persists in ayo data directory
- Restarts on crash
- Logs available via ayo service logs

