---
id: ayo-84rs
status: in_progress
deps: []
links: []
created: 2026-02-06T22:29:46Z
type: bug
priority: 2
assignee: Alex Cabrera
tags: [daemon]
---
# Daemon doesn't handle stale socket file

## Issue
When a stale socket file exists (e.g., from crashed daemon), `ayo daemon start` fails silently.

## Reproduction
1. Kill daemon: `pkill -f 'ayo daemon'`
2. Create stale socket: `touch ~/.local/share/ayo/daemon.sock`
3. Start daemon: `ayo daemon start`
4. Check status: `ayo status` shows 'not running'

## Expected
Daemon should detect stale socket and remove it before starting, or show clear error message.

## Actual
Daemon fails silently, no error message, status shows not running.

## Workaround
Manually remove socket: `rm ~/.local/share/ayo/daemon.sock`


## Notes

**2026-02-06T22:37:32Z**

Investigation: The stale socket file is properly removed (server.go:118). However, there appears to be a startup issue where the daemon process starts but pool.Start() or subsequent steps fail silently, preventing PID file write. This needs deeper investigation. Marking as deferred for now.
