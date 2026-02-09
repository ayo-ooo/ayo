---
id: ase-eyvx
status: closed
deps: [ase-fd17, ase-dws5, ase-8298]
links: []
created: 2026-02-09T03:15:51Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Automated integration tests for critical paths

## Background

Integration tests verify that multiple components work together correctly in realistic scenarios. Unlike unit tests (isolated) and functional tests (single subsystem), integration tests exercise end-to-end paths through the system.

## Why This Matters

Individual components may work perfectly but fail when combined:
- Race conditions between subsystems
- Incorrect data formats at boundaries
- Resource contention
- State management across components

Integration tests catch these issues before they reach users.

## Implementation Details

### Test Structure

```
tests/
  integration/
    main_test.go          # Setup/teardown, test helpers
    flow_execution_test.go
    agent_orchestration_test.go
    trigger_test.go
    chat_test.go
```

### Test Infrastructure

```go
// tests/integration/main_test.go
var (
    testDaemon *DaemonProcess
    testDB     *sql.DB
    cleanup    func()
)

func TestMain(m *testing.M) {
    // Start daemon in test mode
    testDaemon, testDB, cleanup = setupTestEnvironment()
    
    code := m.Run()
    
    cleanup()
    os.Exit(code)
}

func setupTestEnvironment() (*DaemonProcess, *sql.DB, func()) {
    // 1. Create temp directories
    tmpDir := os.MkdirTemp("", "ayo-integration-*")
    
    // 2. Initialize SQLite database
    db := initTestDB(filepath.Join(tmpDir, "test.db"))
    
    // 3. Start daemon with test config
    daemon := startDaemon(&DaemonConfig{
        SocketPath: filepath.Join(tmpDir, "daemon.sock"),
        Database:   db,
        AgentsDir:  filepath.Join(tmpDir, "agents"),
        MatrixMode: "mock", // Use mock Matrix server
    })
    
    // 4. Wait for daemon ready
    waitForDaemon(daemon, 5*time.Second)
    
    cleanup := func() {
        daemon.Stop()
        db.Close()
        os.RemoveAll(tmpDir)
    }
    
    return daemon, db, cleanup
}
```

### Flow Execution Integration Tests

**flow_execution_test.go:**

```go
func TestIntegration_FlowWithAgentStep(t *testing.T) {
    // Create test agent
    createAgent(t, "@test-summarizer", "You summarize text concisely.")
    
    // Create flow that uses agent
    flowContent := `
version: 1
name: integration-test-flow
steps:
  - id: generate
    type: shell
    run: echo "This is a long document that needs summarizing."
  - id: summarize
    type: agent
    agent: "@test-summarizer"
    prompt: "Summarize: {{ steps.generate.stdout }}"
`
    flowPath := createFlowFile(t, "test-flow.yaml", flowContent)
    
    // Execute via CLI
    output, err := runCLI("ayo", "flows", "run", flowPath)
    
    assert.NoError(t, err)
    assert.Contains(t, output, "completed")
    
    // Verify flow_runs recorded
    var runID string
    testDB.QueryRow("SELECT id FROM flow_runs WHERE flow_name = ? ORDER BY started_at DESC LIMIT 1", 
        "integration-test-flow").Scan(&runID)
    assert.NotEmpty(t, runID)
}

func TestIntegration_FlowTriggeredByCron(t *testing.T) {
    // Create flow with immediate trigger (for testing)
    flowContent := `
version: 1
name: cron-trigger-test
steps:
  - id: log
    type: shell
    run: date >> /tmp/cron-test.log
triggers:
  - id: frequent
    type: cron
    schedule: "* * * * *"  # Every minute
`
    createFlowFile(t, "cron-test.yaml", flowContent)
    
    // Wait for trigger to fire
    time.Sleep(65 * time.Second)
    
    // Verify execution
    content, err := os.ReadFile("/tmp/cron-test.log")
    assert.NoError(t, err)
    assert.NotEmpty(t, content)
}

func TestIntegration_FlowParallelSteps(t *testing.T) {
    flowContent := `
version: 1
name: parallel-integration
steps:
  - id: slow1
    type: shell
    run: sleep 2 && echo "done1"
  - id: slow2
    type: shell
    run: sleep 2 && echo "done2"
  - id: combine
    type: shell
    run: echo "{{ steps.slow1.stdout }} and {{ steps.slow2.stdout }}"
    depends_on: [slow1, slow2]
`
    flowPath := createFlowFile(t, "parallel.yaml", flowContent)
    
    start := time.Now()
    output, err := runCLI("ayo", "flows", "run", flowPath)
    elapsed := time.Since(start)
    
    assert.NoError(t, err)
    // Parallel execution should complete in ~2s, not 4s
    assert.Less(t, elapsed, 3*time.Second)
}
```

### Agent Orchestration Integration Tests

**agent_orchestration_test.go:**

```go
func TestIntegration_AyoCreatesAgent(t *testing.T) {
    // Simulate @ayo deciding to create an agent
    output, err := runCLI("ayo", "chat", "@ayo", 
        "I need help with code reviews. Create an agent for that.")
    
    assert.NoError(t, err)
    // @ayo should respond and potentially create agent
    
    // Check if agent was created
    agents, _ := runCLI("ayo", "agents", "list", "--json")
    // Parse and verify
}

func TestIntegration_AgentToAgentChat(t *testing.T) {
    // Create two agents
    createAgent(t, "@analyst", "You analyze data.")
    createAgent(t, "@reporter", "You write reports.")
    
    // Trigger conversation
    output, err := runCLI("ayo", "chat", "@ayo",
        "Have @analyst analyze the sales data and @reporter write a summary.")
    
    assert.NoError(t, err)
    
    // Verify messages exchanged via Matrix
    messages := getMatrixMessages(t, "@analyst")
    assert.NotEmpty(t, messages)
}

func TestIntegration_CapabilitySearch(t *testing.T) {
    // Create agents with known capabilities
    createAgent(t, "@security-expert", "You are a security auditor who finds vulnerabilities.")
    createAgent(t, "@writer", "You write documentation and blog posts.")
    
    // Wait for capability inference
    time.Sleep(2 * time.Second)
    
    // Search for security capability
    output, err := runCLI("ayo", "agents", "capabilities", "--search", "find security bugs")
    
    assert.NoError(t, err)
    assert.Contains(t, output, "@security-expert")
    assert.NotContains(t, output, "@writer")
}
```

### Trigger Integration Tests

**trigger_test.go:**

```go
func TestIntegration_WatchTrigger(t *testing.T) {
    watchDir := t.TempDir()
    
    flowContent := fmt.Sprintf(`
version: 1
name: watch-test
steps:
  - id: log
    type: shell
    run: echo "File changed" >> /tmp/watch-test.log
triggers:
  - id: watcher
    type: watch
    path: %s
    pattern: "*.txt"
`, watchDir)
    createFlowFile(t, "watch-test.yaml", flowContent)
    
    // Give trigger time to register
    time.Sleep(1 * time.Second)
    
    // Create file in watched directory
    os.WriteFile(filepath.Join(watchDir, "test.txt"), []byte("content"), 0644)
    
    // Wait for trigger
    time.Sleep(2 * time.Second)
    
    // Verify triggered
    content, err := os.ReadFile("/tmp/watch-test.log")
    assert.NoError(t, err)
    assert.Contains(t, string(content), "File changed")
}
```

### Chat Integration Tests

**chat_test.go:**

```go
func TestIntegration_UserToAgentChat(t *testing.T) {
    createAgent(t, "@helper", "You are a helpful assistant.")
    
    output, err := runCLI("ayo", "chat", "@helper", "Hello, how are you?")
    
    assert.NoError(t, err)
    assert.NotEmpty(t, output)
    
    // Verify stored in user_messages
    var count int
    testDB.QueryRow("SELECT COUNT(*) FROM user_messages WHERE content LIKE ?", "%Hello%").Scan(&count)
    assert.Equal(t, 1, count)
}

func TestIntegration_ChatHistory(t *testing.T) {
    createAgent(t, "@chat-history-test", "You remember our conversation.")
    
    // First message
    runCLI("ayo", "chat", "@chat-history-test", "My favorite color is blue.")
    
    // Second message referencing first
    output, err := runCLI("ayo", "chat", "@chat-history-test", "What is my favorite color?")
    
    assert.NoError(t, err)
    assert.Contains(t, strings.ToLower(output), "blue")
}
```

### CI/CD Configuration

```yaml
# .github/workflows/integration.yml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    timeout-minutes: 30
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Build
        run: go build -o ayo ./cmd/ayo/...
      
      - name: Run Integration Tests
        run: go test -v -timeout 20m ./tests/integration/...
        env:
          INTEGRATION_TEST: "1"
```

## Acceptance Criteria

- [ ] Test infrastructure with daemon setup/teardown
- [ ] Flow with agent step integration test
- [ ] Flow triggered by cron integration test
- [ ] Parallel steps integration test
- [ ] @ayo agent creation integration test
- [ ] Agent-to-agent chat integration test
- [ ] Capability search integration test
- [ ] Watch trigger integration test
- [ ] User-to-agent chat integration test
- [ ] Chat history integration test
- [ ] CI/CD workflow for integration tests
- [ ] All tests pass in under 20 minutes
- [ ] Tests are idempotent (can run multiple times)

