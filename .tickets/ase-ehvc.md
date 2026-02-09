---
id: ase-ehvc
status: open
deps: [ase-0hi0]
links: []
created: 2026-02-09T03:27:49Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for flows CLI commands

## Background

The flows CLI (ase-0hi0) provides flow management and execution. These commands need unit tests.

## Test Cases

### flows list Tests

```go
func TestFlowsList_Empty(t *testing.T) {
    output, err := runCmd("flows", "list")
    assert.NoError(t, err)
    assert.Contains(t, output, "No flows")
}

func TestFlowsList_WithFlows(t *testing.T) {
    setupTestFlows(t, "daily-digest", "process-data")
    output, err := runCmd("flows", "list")
    assert.NoError(t, err)
    assert.Contains(t, output, "daily-digest")
    assert.Contains(t, output, "process-data")
}

func TestFlowsList_JSON(t *testing.T) {
    setupTestFlows(t, "test-flow")
    output, err := runCmd("flows", "list", "--json")
    assert.NoError(t, err)
    var flows []FlowSummary
    err = json.Unmarshal([]byte(output), &flows)
    assert.NoError(t, err)
    assert.Len(t, flows, 1)
}
```

### flows show Tests

```go
func TestFlowsShow_Basic(t *testing.T) {
    setupTestFlow(t, "my-flow")
    output, err := runCmd("flows", "show", "my-flow")
    assert.NoError(t, err)
    assert.Contains(t, output, "my-flow")
    assert.Contains(t, output, "steps:")
}

func TestFlowsShow_NotFound(t *testing.T) {
    _, err := runCmd("flows", "show", "nonexistent")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}

func TestFlowsShow_JSON(t *testing.T) {
    setupTestFlow(t, "my-flow")
    output, err := runCmd("flows", "show", "my-flow", "--json")
    assert.NoError(t, err)
    var flow Flow
    err = json.Unmarshal([]byte(output), &flow)
    assert.NoError(t, err)
}
```

### flows run Tests

```go
func TestFlowsRun_Simple(t *testing.T) {
    createTestFlow(t, "simple", `
version: 1
name: simple
steps:
  - id: hello
    type: shell
    run: echo "Hello"
`)
    output, err := runCmd("flows", "run", "simple")
    assert.NoError(t, err)
    assert.Contains(t, output, "Hello")
}

func TestFlowsRun_WithParams(t *testing.T) {
    createTestFlow(t, "parameterized", `
version: 1
name: parameterized
steps:
  - id: greet
    type: shell
    run: echo "Hello {{ params.name }}"
`)
    output, err := runCmd("flows", "run", "parameterized", "-p", "name=World")
    assert.NoError(t, err)
    assert.Contains(t, output, "Hello World")
}

func TestFlowsRun_WithInputFile(t *testing.T) {
    createTestFlow(t, "input-flow", `...`)
    inputFile := createTempJSON(t, map[string]any{"key": "value"})
    output, err := runCmd("flows", "run", "input-flow", "-i", inputFile)
    assert.NoError(t, err)
}

func TestFlowsRun_NotFound(t *testing.T) {
    _, err := runCmd("flows", "run", "nonexistent")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}

func TestFlowsRun_InvalidFlow(t *testing.T) {
    createTestFlow(t, "invalid", "not valid yaml: {{")
    _, err := runCmd("flows", "run", "invalid")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "parse")
}

func TestFlowsRun_Async(t *testing.T) {
    createTestFlow(t, "slow", `
version: 1
name: slow
steps:
  - id: wait
    type: shell
    run: sleep 10
`)
    output, err := runCmd("flows", "run", "slow", "--async")
    assert.NoError(t, err)
    assert.Contains(t, output, "run-id")
    // Should return immediately
}

func TestFlowsRun_JSON(t *testing.T) {
    createTestFlow(t, "json-test", `...`)
    output, err := runCmd("flows", "run", "json-test", "--json")
    assert.NoError(t, err)
    var result FlowRunResult
    err = json.Unmarshal([]byte(output), &result)
    assert.NoError(t, err)
}
```

### flows delete Tests

```go
func TestFlowsDelete_Basic(t *testing.T) {
    createTestFlow(t, "to-delete", `...`)
    output, err := runCmd("flows", "delete", "to-delete", "-f")
    assert.NoError(t, err)
    assert.Contains(t, output, "Deleted")
    
    // Verify gone
    _, err = runCmd("flows", "show", "to-delete")
    assert.Error(t, err)
}

func TestFlowsDelete_NotFound(t *testing.T) {
    _, err := runCmd("flows", "delete", "nonexistent", "-f")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}

func TestFlowsDelete_RequiresConfirmation(t *testing.T) {
    createTestFlow(t, "protected", `...`)
    // Without -f, should prompt (or error in non-interactive)
    _, err := runCmd("flows", "delete", "protected")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "confirm")
}
```

### flows history Tests

```go
func TestFlowsHistory_Empty(t *testing.T) {
    output, err := runCmd("flows", "history")
    assert.NoError(t, err)
    assert.Contains(t, output, "No runs")
}

func TestFlowsHistory_WithRuns(t *testing.T) {
    setupTestFlowRuns(t, 5)
    output, err := runCmd("flows", "history")
    assert.NoError(t, err)
    // Should show runs
}

func TestFlowsHistory_Limit(t *testing.T) {
    setupTestFlowRuns(t, 50)
    output, err := runCmd("flows", "history", "--limit", "10")
    assert.NoError(t, err)
    // Verify limited
}

func TestFlowsHistory_FilterByFlow(t *testing.T) {
    setupTestFlowRuns(t, 10) // Mixed flows
    output, err := runCmd("flows", "history", "--flow", "daily-digest")
    assert.NoError(t, err)
    // Only daily-digest runs
}
```

### Files to Create

1. `cmd/ayo/flows_test.go` - All flows CLI tests

## Acceptance Criteria

- [ ] list subcommand tests (empty, with flows, JSON)
- [ ] show subcommand tests (basic, not found, JSON)
- [ ] run subcommand tests (simple, params, input, async, errors)
- [ ] delete subcommand tests (basic, confirmation, not found)
- [ ] history subcommand tests (empty, with runs, filters)
- [ ] Error message verification
- [ ] JSON output parsing
- [ ] All tests pass

