---
id: ase-7pye
status: closed
deps: [ase-t4cr]
links: []
created: 2026-02-06T04:15:20Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-rzhr
---
# Integration tests for daemon

Add integration tests for daemon functionality including triggers and session management.

## Design

## Test Coverage
1. Daemon start/stop lifecycle
2. Sandbox auto-start
3. Trigger registration and firing
4. Session management
5. IRC bridge functionality

## Test Structure
internal/integration/daemon_test.go

## Tests
func TestDaemon_Lifecycle(t *testing.T)
  - Start daemon
  - Verify sandbox running
  - Stop daemon
  - Verify clean shutdown

func TestDaemon_TriggerCron(t *testing.T)
  - Register cron trigger (every second for test)
  - Wait for trigger fire
  - Verify agent session started
  - Cleanup

func TestDaemon_TriggerWatch(t *testing.T)
  - Register file watch trigger
  - Modify watched file
  - Verify trigger fires
  - Cleanup

func TestDaemon_SessionManagement(t *testing.T)
  - Start session via daemon
  - List sessions
  - Stop session
  - Verify cleanup

## Mock Options
For unit tests, use mock providers.
For integration tests, use real sandbox if available.

## Acceptance Criteria

- Daemon lifecycle tested
- Triggers tested end-to-end
- Session management tested
- Mock fallback for CI

