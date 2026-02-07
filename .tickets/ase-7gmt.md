---
id: ase-7gmt
status: closed
deps: [ase-2msm]
links: []
created: 2026-02-06T04:15:12Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-ka3q
---
# Integration tests for Alpine sandbox

Add integration tests for the Alpine-based sandbox with IRC.

## Design

## Test Coverage
1. Sandbox creation with Alpine image
2. Agent user creation
3. IRC server startup and connectivity
4. Directory structure verification
5. Inter-agent messaging via IRC

## Test Structure
internal/integration/sandbox_alpine_test.go

## Tests
func TestAlpineSandbox_Create(t *testing.T)
  - Create sandbox with Alpine
  - Verify image is Alpine
  - Verify basic commands work

func TestAlpineSandbox_UserCreation(t *testing.T)
  - Create agent user
  - Verify home directory exists
  - Verify user can execute commands

func TestAlpineSandbox_IRC(t *testing.T)
  - Verify ngircd is running
  - Send message via nc
  - Verify message in logs

func TestAlpineSandbox_InterAgentMessage(t *testing.T)
  - Create two agent users
  - Agent A sends message to #general
  - Verify message appears in logs

## Skip Conditions
Skip if:
- Apple Container not available (macOS)
- systemd-nspawn not available (Linux)
- CI environment without containers

## Acceptance Criteria

- All tests pass on supported platforms
- Tests skip gracefully on unsupported platforms
- IRC messaging tested end-to-end

