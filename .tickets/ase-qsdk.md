---
id: ase-qsdk
status: open
deps: [ase-trs4, ase-8g9z]
links: []
created: 2026-02-09T03:26:06Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for trigger CLI commands

## Background

The Trigger CLI Redesign (ase-6khq) adds new subcommands: schedule, watch, and removes the generic add command. These CLI changes need unit tests.

## Why This Matters

Trigger CLI is user-facing and handles time-sensitive operations. Tests ensure:
- Commands parse arguments correctly
- Validation catches invalid inputs
- Error messages are helpful
- JSON output is correct

## Test Cases

### trigger schedule Tests

```go
func TestTriggerSchedule_Basic(t *testing.T) {
    output, err := runCmd("trigger", "schedule", "@backup", "0 2 * * *")
    assert.NoError(t, err)
    assert.Contains(t, output, "Trigger created")
}

func TestTriggerSchedule_WithPrompt(t *testing.T) {
    output, err := runCmd("trigger", "schedule", "@backup", "0 2 * * *", 
        "--prompt", "Run nightly backup")
    assert.NoError(t, err)
    // Verify prompt stored
}

func TestTriggerSchedule_InvalidCron(t *testing.T) {
    _, err := runCmd("trigger", "schedule", "@backup", "invalid")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid cron")
}

func TestTriggerSchedule_MissingAgent(t *testing.T) {
    _, err := runCmd("trigger", "schedule", "0 * * * *")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "agent required")
}

func TestTriggerSchedule_UnknownAgent(t *testing.T) {
    _, err := runCmd("trigger", "schedule", "@nonexistent", "0 * * * *")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "agent not found")
}

func TestTriggerSchedule_NaturalLanguage(t *testing.T) {
    // After natural language parsing is implemented
    output, err := runCmd("trigger", "schedule", "@backup", "every day at 2am")
    assert.NoError(t, err)
}
```

### trigger watch Tests

```go
func TestTriggerWatch_Basic(t *testing.T) {
    tmpDir := t.TempDir()
    output, err := runCmd("trigger", "watch", tmpDir, "@build")
    assert.NoError(t, err)
    assert.Contains(t, output, "Trigger created")
}

func TestTriggerWatch_WithPatterns(t *testing.T) {
    tmpDir := t.TempDir()
    output, err := runCmd("trigger", "watch", tmpDir, "@build", "*.go", "*.mod")
    assert.NoError(t, err)
}

func TestTriggerWatch_Recursive(t *testing.T) {
    tmpDir := t.TempDir()
    output, err := runCmd("trigger", "watch", tmpDir, "@build", "--recursive")
    assert.NoError(t, err)
}

func TestTriggerWatch_InvalidPath(t *testing.T) {
    _, err := runCmd("trigger", "watch", "/nonexistent/path", "@build")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "path not found")
}

func TestTriggerWatch_RelativePath(t *testing.T) {
    // Should convert to absolute
    output, err := runCmd("trigger", "watch", "./src", "@build")
    assert.NoError(t, err)
}
```

### trigger list Tests

```go
func TestTriggerList_Empty(t *testing.T) {
    output, err := runCmd("trigger", "list")
    assert.NoError(t, err)
    assert.Contains(t, output, "No triggers")
}

func TestTriggerList_WithTriggers(t *testing.T) {
    // Setup triggers first
    setupTestTriggers(t)
    
    output, err := runCmd("trigger", "list")
    assert.NoError(t, err)
    assert.Contains(t, output, "backup-daily")
}

func TestTriggerList_JSON(t *testing.T) {
    setupTestTriggers(t)
    
    output, err := runCmd("trigger", "list", "--json")
    assert.NoError(t, err)
    
    var triggers []Trigger
    err = json.Unmarshal([]byte(output), &triggers)
    assert.NoError(t, err)
}
```

### trigger rm Tests

```go
func TestTriggerRm_ByID(t *testing.T) {
    id := setupTestTrigger(t)
    output, err := runCmd("trigger", "rm", id)
    assert.NoError(t, err)
    assert.Contains(t, output, "Removed")
}

func TestTriggerRm_ByPrefix(t *testing.T) {
    setupTestTrigger(t)  // Creates trigger with known prefix
    output, err := runCmd("trigger", "rm", "back")  // Prefix match
    assert.NoError(t, err)
}

func TestTriggerRm_AmbiguousPrefix(t *testing.T) {
    setupMultipleTriggers(t)  // backup-daily, backup-weekly
    _, err := runCmd("trigger", "rm", "backup")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "ambiguous")
}
```

### Files to Create

1. `cmd/ayo/trigger_test.go` - All trigger CLI tests

### Test Infrastructure

```go
func runCmd(args ...string) (string, error) {
    // Create command with test daemon
    // Capture output
    // Return
}

func setupTestDaemon(t *testing.T) *TestDaemon {
    // Start daemon in test mode
    // Return for cleanup
}
```

## Acceptance Criteria

- [ ] schedule subcommand tests (valid, invalid cron, missing args)
- [ ] watch subcommand tests (valid, patterns, recursive, invalid path)
- [ ] list subcommand tests (empty, with triggers, JSON output)
- [ ] rm subcommand tests (by ID, by prefix, ambiguous)
- [ ] show subcommand tests
- [ ] Error messages verified in tests
- [ ] JSON output parsing tests
- [ ] All tests pass

