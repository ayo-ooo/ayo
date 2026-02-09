---
id: ase-dws5
status: closed
deps: [ase-euxv]
links: []
created: 2026-02-09T03:14:35Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Functional tests for flow execution

## Background

The flow executor runs multi-step workflows defined in YAML. It handles shell commands, agent invocations, template substitution, and step dependencies.

## Why This Matters

Flow execution is complex with many failure modes:
- Step dependency resolution
- Template substitution timing
- Agent invocation and response handling
- Error propagation and recovery
- Concurrent step execution

Functional tests verify the executor handles all these correctly.

## Implementation Details

### Test Structure

```
internal/
  flows/
    executor.go
    executor_test.go
    testdata/
      flows/
        simple-shell.yaml
        multi-step.yaml
        agent-step.yaml
        parallel-steps.yaml
        conditional-flow.yaml
        error-recovery.yaml
```

### Test Helpers

```go
// internal/flows/executor_test.go
func setupTestExecutor(t *testing.T) (*Executor, *MockAgentInvoker) {
    invoker := NewMockAgentInvoker()
    executor := NewExecutor(
        WithAgentInvoker(invoker),
        WithWorkDir(t.TempDir()),
    )
    return executor, invoker
}

type MockAgentInvoker struct {
    responses map[string]string  // agent -> response
    calls     []AgentCall
}

func (m *MockAgentInvoker) Invoke(agent, prompt string) (string, error) {
    m.calls = append(m.calls, AgentCall{agent, prompt})
    return m.responses[agent], nil
}
```

### Test Cases

**executor_test.go:**

```go
func TestExecutor_SimpleShell(t *testing.T) {
    flow := parseFlow(t, `
version: 1
name: simple
steps:
  - id: hello
    type: shell
    run: echo "Hello World"
`)
    
    result, err := executor.Execute(flow)
    
    assert.NoError(t, err)
    assert.Equal(t, "Hello World\n", result.Steps["hello"].Stdout)
}

func TestExecutor_MultiStep(t *testing.T) {
    flow := parseFlow(t, `
version: 1
name: multi
steps:
  - id: list
    type: shell
    run: echo -e "file1\nfile2\nfile3"
  - id: count
    type: shell
    run: echo "{{ steps.list.stdout }}" | wc -l
`)
    
    result, err := executor.Execute(flow)
    
    assert.NoError(t, err)
    assert.Contains(t, result.Steps["count"].Stdout, "3")
}

func TestExecutor_AgentStep(t *testing.T) {
    invoker.SetResponse("@summarizer", "Summary: This is a summary")
    
    flow := parseFlow(t, `
version: 1
name: with-agent
steps:
  - id: summarize
    type: agent
    agent: "@summarizer"
    prompt: "Summarize this document"
    input: "Long document content here..."
`)
    
    result, err := executor.Execute(flow)
    
    assert.NoError(t, err)
    assert.Contains(t, result.Steps["summarize"].Output, "Summary")
    assert.Len(t, invoker.Calls(), 1)
}

func TestExecutor_TemplateSubstitution(t *testing.T) {
    invoker.SetResponse("@analyzer", "Found 5 issues")
    
    flow := parseFlow(t, `
version: 1
name: template
steps:
  - id: gather
    type: shell
    run: find . -name "*.go"
  - id: analyze
    type: agent
    agent: "@analyzer"
    prompt: "Analyze these files:\n{{ steps.gather.stdout }}"
`)
    
    _, err := executor.Execute(flow)
    
    assert.NoError(t, err)
    call := invoker.LastCall()
    assert.Contains(t, call.Prompt, "Analyze these files:")
}

func TestExecutor_ParallelSteps(t *testing.T) {
    flow := parseFlow(t, `
version: 1
name: parallel
steps:
  - id: slow1
    type: shell
    run: sleep 1 && echo "done1"
  - id: slow2
    type: shell
    run: sleep 1 && echo "done2"
  - id: combine
    type: shell
    run: echo "{{ steps.slow1.stdout }} {{ steps.slow2.stdout }}"
    depends_on: [slow1, slow2]
`)
    
    start := time.Now()
    result, err := executor.Execute(flow)
    elapsed := time.Since(start)
    
    assert.NoError(t, err)
    // Should run in ~1 second (parallel), not 2 seconds (sequential)
    assert.Less(t, elapsed, 1500*time.Millisecond)
}

func TestExecutor_DependencyOrder(t *testing.T) {
    flow := parseFlow(t, `
version: 1
name: deps
steps:
  - id: step3
    type: shell
    run: echo "step3"
    depends_on: [step2]
  - id: step1
    type: shell
    run: echo "step1"
  - id: step2
    type: shell
    run: echo "step2 after {{ steps.step1.stdout }}"
    depends_on: [step1]
`)
    
    result, err := executor.Execute(flow)
    
    assert.NoError(t, err)
    // Verify execution order via timestamps
    assert.True(t, result.Steps["step1"].EndTime.Before(result.Steps["step2"].StartTime))
    assert.True(t, result.Steps["step2"].EndTime.Before(result.Steps["step3"].StartTime))
}

func TestExecutor_ShellError(t *testing.T) {
    flow := parseFlow(t, `
version: 1
name: error
steps:
  - id: fail
    type: shell
    run: exit 1
  - id: after
    type: shell
    run: echo "should not run"
    depends_on: [fail]
`)
    
    result, err := executor.Execute(flow)
    
    assert.Error(t, err)
    assert.Equal(t, 1, result.Steps["fail"].ExitCode)
    assert.Nil(t, result.Steps["after"]) // Not executed
}

func TestExecutor_ContinueOnError(t *testing.T) {
    flow := parseFlow(t, `
version: 1
name: continue
steps:
  - id: fail
    type: shell
    run: exit 1
    continue_on_error: true
  - id: after
    type: shell
    run: echo "still running"
    depends_on: [fail]
`)
    
    result, err := executor.Execute(flow)
    
    assert.NoError(t, err) // Overall success
    assert.Equal(t, 1, result.Steps["fail"].ExitCode)
    assert.NotNil(t, result.Steps["after"]) // Still executed
}

func TestExecutor_Timeout(t *testing.T) {
    flow := parseFlow(t, `
version: 1
name: timeout
steps:
  - id: slow
    type: shell
    run: sleep 10
    timeout: 1s
`)
    
    result, err := executor.Execute(flow)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}

func TestExecutor_WorkingDirectory(t *testing.T) {
    tmpDir := t.TempDir()
    os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644)
    
    executor := NewExecutor(WithWorkDir(tmpDir))
    flow := parseFlow(t, `
version: 1
name: workdir
steps:
  - id: read
    type: shell
    run: cat test.txt
`)
    
    result, err := executor.Execute(flow)
    
    assert.NoError(t, err)
    assert.Equal(t, "content", result.Steps["read"].Stdout)
}

func TestExecutor_EnvironmentVariables(t *testing.T) {
    flow := parseFlow(t, `
version: 1
name: env
steps:
  - id: check
    type: shell
    run: echo $FLOW_NAME
    env:
      FLOW_NAME: my-flow
`)
    
    result, err := executor.Execute(flow)
    
    assert.NoError(t, err)
    assert.Contains(t, result.Steps["check"].Stdout, "my-flow")
}

func TestExecutor_RecordsToDatabase(t *testing.T) {
    db := setupTestDB(t)
    executor := NewExecutor(WithDatabase(db))
    
    flow := parseFlow(t, `
version: 1
name: recorded
steps:
  - id: hello
    type: shell
    run: echo "hello"
`)
    
    result, err := executor.Execute(flow)
    
    assert.NoError(t, err)
    
    // Verify recorded in flow_runs table
    var count int
    db.QueryRow("SELECT COUNT(*) FROM flow_runs WHERE flow_name = ?", "recorded").Scan(&count)
    assert.Equal(t, 1, count)
}
```

## Acceptance Criteria

- [ ] Simple shell step execution
- [ ] Multi-step with template substitution
- [ ] Agent step invocation
- [ ] Parallel step execution
- [ ] Dependency ordering respected
- [ ] Shell error stops dependent steps
- [ ] continue_on_error works
- [ ] Step timeout handling
- [ ] Working directory configuration
- [ ] Environment variable injection
- [ ] Execution recorded to database
- [ ] All tests pass with -race flag
- [ ] Tests complete in < 30 seconds

