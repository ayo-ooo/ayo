---
id: ayo-eude
status: closed
deps: []
links: []
created: 2026-02-06T22:24:31Z
type: bug
priority: 1
assignee: Alex Cabrera
tags: [daemon, cli]
---
# 'ayo daemon stop' does not kill daemon process

## Issue
`ayo daemon stop` reports 'Daemon stopped' but the process continues running.

## Reproduction
1. Start daemon: `ayo daemon start`
2. Verify running: `pgrep -f 'ayo daemon'` shows PID
3. Stop daemon: `ayo daemon stop` - says 'Daemon stopped'
4. Check process: `pgrep -f 'ayo daemon'` - STILL shows same PID
5. Status shows not running (socket removed but process alive)

## Expected
Process should be terminated when stop is called.

## Actual
- Socket file is removed
- 'Daemon stopped' message shown
- Process continues running
- Requires manual `kill <pid>` to actually stop

## Impact
Daemon processes accumulate, socket conflicts on restart.


## Notes

**2026-02-06T22:34:16Z**

FIXED: Added os.Exit(0) after s.Stop() in handleShutdown() in internal/daemon/server.go. Daemon now properly terminates after shutdown RPC.
