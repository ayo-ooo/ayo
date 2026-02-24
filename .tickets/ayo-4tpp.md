---
id: ayo-4tpp
status: closed
deps: []
links: []
created: 2026-02-23T22:16:35Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-i2qo
tags: [testing]
---
# Add comprehensive test coverage

Ensure test coverage exceeds 70% for core packages. Add integration tests for key workflows.

## Context

Before GTM, the codebase needs solid test coverage to prevent regressions and enable confident refactoring.

## Target Packages

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| internal/sandbox | ~40% | 70% | High |
| internal/squads | ~30% | 70% | High |
| internal/planners | ~50% | 70% | High |
| internal/daemon/triggers | ~30% | 70% | High |
| internal/tickets | ~60% | 70% | Medium |
| internal/tools | ~50% | 70% | Medium |
| internal/agent | ~45% | 70% | Medium |

## Unit Tests to Add

### Sandbox Package

```go
// sandbox_test.go
func TestCreateSandbox(t *testing.T)
func TestDestroySandbox(t *testing.T)
func TestExecInSandbox(t *testing.T)
func TestSandboxIsolation(t *testing.T)
```

### Squads Package

```go
// squads_test.go
func TestCreateSquad(t *testing.T)
func TestLoadConstitution(t *testing.T)
func TestDispatchRouting(t *testing.T)
func TestAgentUserCreation(t *testing.T)
```

### Triggers Package

```go
// trigger_engine_test.go
func TestCronTrigger(t *testing.T)
func TestIntervalTrigger(t *testing.T)
func TestOneTimeTrigger(t *testing.T)
func TestFileWatchTrigger(t *testing.T)
func TestTriggerPersistence(t *testing.T)
```

### Planners Package

```go
// planners_test.go
func TestLoadPlanner(t *testing.T)
func TestTodosPlanner(t *testing.T)
func TestTicketsPlanner(t *testing.T)
```

## Integration Tests

```go
// internal/integration/
func TestAgentFullWorkflow(t *testing.T) {
    // Create agent → Run prompt → Verify output
}

func TestSquadWorkflow(t *testing.T) {
    // Create squad → Dispatch → Verify routing → Check tickets
}

func TestTriggerExecution(t *testing.T) {
    // Create trigger → Wait for fire → Verify agent ran
}

func TestFileRequestApproval(t *testing.T) {
    // Agent writes file → Approval flow → Verify on host
}
```

## Test Helpers

Create shared test utilities:

```go
// internal/testutil/
func NewTestSandbox(t *testing.T) *Sandbox
func NewTestSquad(t *testing.T, config SquadConfig) *Squad
func NewTestAgent(t *testing.T, name string) *Agent
func WaitForTrigger(t *testing.T, name string, timeout time.Duration)
```

## CI Integration

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: go test ./... -count=1 -race -coverprofile=coverage.out

- name: Check coverage
  run: |
    go tool cover -func=coverage.out | grep total | awk '{print $3}'
    # Fail if under 70%
```

## Files to Create/Modify

1. **`internal/sandbox/*_test.go`** - Sandbox tests
2. **`internal/squads/*_test.go`** - Squad tests
3. **`internal/daemon/triggers/*_test.go`** - Trigger tests
4. **`internal/planners/*_test.go`** - Planner tests
5. **`internal/testutil/`** (new) - Test utilities
6. **`internal/integration/`** (new) - Integration tests

## Acceptance Criteria

- [ ] Each target package has 70%+ coverage
- [ ] Integration tests for key workflows pass
- [ ] Tests run in CI on every PR
- [ ] Coverage reported in CI
- [ ] Test helpers documented
- [ ] No flaky tests

## Measuring Coverage

```bash
# Full coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Per-package coverage
go test ./internal/sandbox/... -cover
go test ./internal/squads/... -cover
```
