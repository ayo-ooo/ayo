---
id: ayo-rx11
status: closed
deps: []
links: []
created: 2026-02-24T03:00:00Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-rx10
tags: [remediation, testing]
---
# Task: Sandbox Package Tests (70% Coverage)

## Summary

Increase `internal/sandbox` test coverage from 26.9% to 70%.

## Final State

```bash
go test ./internal/sandbox/... -cover
# sandbox: 30.5% of statements
# workingcopy: 75.8% of statements
# combined: 35.7%
```

## Tests Added

- Pool tests: Config, ReleaseAgent, GetSandboxAgents, List, AcquireWithOptions (JoinSandbox, Group)
- Squad unit test: SquadSandboxName
- NoneProvider tests: EnsureAgentUser, Stats
- WorkingCopy tests: Sync, Diff, extractTar, CreateTarFromDir, edge cases

## Blockers for 70%

The remaining untested functions (apple.go, linux.go, squad.go, ayo.go, bootstrap.go) 
require actual container runtimes which are not available in unit tests:
- Apple Container operations (Create, Exec, Start, Stop, etc.)
- Linux nspawn operations
- ayod protocol testing
- Squad sandbox creation

These would require integration tests with build tags for container environments.

## Current State

```bash
go test ./internal/sandbox/... -cover
# coverage: 26.9% of statements
```

## Tests to Add

### Core Sandbox Operations

```go
// sandbox_test.go
func TestCreateSandbox(t *testing.T) {
    // Test sandbox creation with various configs
}

func TestDestroySandbox(t *testing.T) {
    // Test sandbox destruction
}

func TestSandboxExists(t *testing.T) {
    // Test existence check
}

func TestSandboxList(t *testing.T) {
    // Test listing sandboxes
}
```

### Execution

```go
// exec_test.go
func TestExecInSandbox(t *testing.T) {
    // Test command execution
}

func TestExecWithEnv(t *testing.T) {
    // Test with custom environment
}

func TestExecTimeout(t *testing.T) {
    // Test timeout handling
}

func TestExecStdin(t *testing.T) {
    // Test stdin handling
}
```

### File Operations

```go
// file_test.go
func TestCopyToSandbox(t *testing.T) {
    // Test file copy in
}

func TestCopyFromSandbox(t *testing.T) {
    // Test file copy out
}

func TestSandboxFileExists(t *testing.T) {
    // Test file existence
}
```

### Provider Interface

```go
// provider_test.go
func TestAppleContainerProvider(t *testing.T) {
    // Test Apple Container if available
}

func TestNspawnProvider(t *testing.T) {
    // Test nspawn if available
}

func TestDummyProvider(t *testing.T) {
    // Test dummy provider
}
```

### Bootstrap

```go
// bootstrap_test.go
func TestBootstrapSquadSandbox(t *testing.T) {
    // Test squad sandbox bootstrap
}

func TestAyodInstallation(t *testing.T) {
    // Test ayod binary installation
}
```

## Strategy

1. Start with mock-based unit tests for core logic
2. Add integration tests where sandbox actually needed
3. Use build tags for provider-specific tests
4. Add table-driven tests for edge cases

## Acceptance Criteria

- [ ] Coverage ≥ 70%
- [ ] All tests pass
- [ ] Tests run in CI (mocked for non-container environments)
- [ ] No flaky tests
