---
id: ase-3jce
status: open
deps: [ase-ip4t]
links: []
created: 2026-02-09T03:28:14Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for agents CLI updates

## Background

The agents CLI is being updated (ase-ip4t) with trust levels, creation metadata, and capabilities display. These changes need unit tests.

## Test Cases

### agents list Tests

```go
func TestAgentsList_ShowsTrustLevel(t *testing.T) {
    setupAgentsWithTrust(t, map[string]string{
        "@safe": "sandboxed",
        "@power": "privileged",
        "@unsafe": "unrestricted",
    })
    output, err := runCmd("agents", "list")
    assert.NoError(t, err)
    assert.Contains(t, output, "sandboxed")
    assert.Contains(t, output, "privileged")
    assert.Contains(t, output, "unrestricted")
}

func TestAgentsList_ShowsCreatedBy(t *testing.T) {
    setupAgentCreatedBy(t, "@doc-writer", "@ayo")
    output, err := runCmd("agents", "list")
    assert.NoError(t, err)
    assert.Contains(t, output, "@ayo")
}

func TestAgentsList_FilterByTrust(t *testing.T) {
    setupAgentsWithTrust(t, map[string]string{
        "@safe1": "sandboxed",
        "@safe2": "sandboxed",
        "@power": "privileged",
    })
    output, err := runCmd("agents", "list", "--trust", "sandboxed")
    assert.NoError(t, err)
    assert.Contains(t, output, "@safe1")
    assert.Contains(t, output, "@safe2")
    assert.NotContains(t, output, "@power")
}

func TestAgentsList_FilterByCreator(t *testing.T) {
    setupAgentCreatedBy(t, "@doc-writer", "@ayo")
    setupAgentCreatedBy(t, "@manual", "user")
    output, err := runCmd("agents", "list", "--created-by", "@ayo")
    assert.NoError(t, err)
    assert.Contains(t, output, "@doc-writer")
    assert.NotContains(t, output, "@manual")
}

func TestAgentsList_JSON(t *testing.T) {
    setupTestAgents(t)
    output, err := runCmd("agents", "list", "--json")
    assert.NoError(t, err)
    var agents []AgentSummary
    err = json.Unmarshal([]byte(output), &agents)
    assert.NoError(t, err)
    // Verify JSON has trust_level and created_by
    for _, a := range agents {
        assert.NotEmpty(t, a.TrustLevel)
    }
}
```

### agents show Tests

```go
func TestAgentsShow_IncludesTrustLevel(t *testing.T) {
    setupAgentWithTrust(t, "@test", "privileged")
    output, err := runCmd("agents", "show", "@test")
    assert.NoError(t, err)
    assert.Contains(t, output, "Trust Level: privileged")
}

func TestAgentsShow_IncludesCreationMetadata(t *testing.T) {
    setupAgentCreatedBy(t, "@doc-writer", "@ayo")
    output, err := runCmd("agents", "show", "@doc-writer")
    assert.NoError(t, err)
    assert.Contains(t, output, "Created By: @ayo")
    assert.Contains(t, output, "Created At:")
}

func TestAgentsShow_IncludesRefinementHistory(t *testing.T) {
    setupAgentWithRefinements(t, "@refined", 3)
    output, err := runCmd("agents", "show", "@refined")
    assert.NoError(t, err)
    assert.Contains(t, output, "Refinements: 3")
}

func TestAgentsShow_IncludesCapabilities(t *testing.T) {
    setupAgentWithCapabilities(t, "@code-reviewer", []string{"code-review", "bug-detection"})
    output, err := runCmd("agents", "show", "@code-reviewer")
    assert.NoError(t, err)
    assert.Contains(t, output, "Capabilities")
    assert.Contains(t, output, "code-review")
}

func TestAgentsShow_JSON(t *testing.T) {
    setupFullAgent(t, "@complete")
    output, err := runCmd("agents", "show", "@complete", "--json")
    assert.NoError(t, err)
    var agent AgentDetails
    err = json.Unmarshal([]byte(output), &agent)
    assert.NoError(t, err)
    assert.NotEmpty(t, agent.TrustLevel)
    assert.NotEmpty(t, agent.CreatedBy)
    assert.NotNil(t, agent.Capabilities)
}
```

### agents capabilities Tests

```go
func TestAgentsCapabilities_Basic(t *testing.T) {
    setupAgentWithCapabilities(t, "@reviewer", []string{"code-review"})
    output, err := runCmd("agents", "capabilities", "@reviewer")
    assert.NoError(t, err)
    assert.Contains(t, output, "code-review")
}

func TestAgentsCapabilities_Search(t *testing.T) {
    setupAgentsWithCapabilities(t, map[string][]string{
        "@security": {"security-audit", "vulnerability-detection"},
        "@writer": {"documentation", "technical-writing"},
    })
    output, err := runCmd("agents", "capabilities", "--search", "security issues")
    assert.NoError(t, err)
    assert.Contains(t, output, "@security")
    // @security should rank higher than @writer
}

func TestAgentsCapabilities_Refresh(t *testing.T) {
    setupAgent(t, "@to-refresh")
    output, err := runCmd("agents", "capabilities", "refresh", "@to-refresh")
    assert.NoError(t, err)
    assert.Contains(t, output, "Refreshed")
}

func TestAgentsCapabilities_All(t *testing.T) {
    setupTestAgents(t)
    output, err := runCmd("agents", "capabilities", "--all")
    assert.NoError(t, err)
    // Should list capabilities for all agents
}
```

### agents create Tests (internal)

```go
func TestAgentsCreate_Basic(t *testing.T) {
    output, err := runCmd("agents", "create",
        "--name", "test-agent",
        "--system-prompt", "You are a test agent",
        "--created-by", "@ayo")
    assert.NoError(t, err)
    assert.Contains(t, output, "Created")
    
    // Verify exists
    _, err = runCmd("agents", "show", "@test-agent")
    assert.NoError(t, err)
}

func TestAgentsCreate_InvalidName(t *testing.T) {
    _, err := runCmd("agents", "create",
        "--name", "has.dot",
        "--system-prompt", "Test",
        "--created-by", "@ayo")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid")
}

func TestAgentsCreate_DuplicateName(t *testing.T) {
    setupAgent(t, "@existing")
    _, err := runCmd("agents", "create",
        "--name", "existing",
        "--system-prompt", "Test",
        "--created-by", "@ayo")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "exists")
}
```

### Files to Create

1. `cmd/ayo/agents_test.go` - All agents CLI tests

## Acceptance Criteria

- [ ] list tests show trust level column
- [ ] list tests show created by column
- [ ] list filter by trust works
- [ ] list filter by creator works
- [ ] show includes all new metadata
- [ ] show includes capabilities summary
- [ ] capabilities subcommand tests
- [ ] capabilities search tests
- [ ] create subcommand tests (internal)
- [ ] JSON output includes all fields
- [ ] All tests pass

