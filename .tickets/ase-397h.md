---
id: ase-397h
status: open
deps: [ase-jep0]
links: []
created: 2026-02-09T03:11:46Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for SQLite repositories

## Background

The Agent Orchestration System introduces several new SQLite tables and repositories:
- user_sessions / user_messages
- flow_runs
- ayo_created_agents
- trigger_stats
- agent_capabilities

Each repository needs comprehensive unit tests to verify CRUD operations, edge cases, and data integrity.

## Why This Matters

Unit tests for repositories:
- Catch bugs early before integration
- Document expected behavior
- Enable safe refactoring
- Verify SQL queries are correct

## Implementation Details

### Test Structure

Create `*_test.go` files alongside each repository:

```
internal/
  repositories/
    user_sessions.go
    user_sessions_test.go
    flow_runs.go
    flow_runs_test.go
    ayo_created_agents.go
    ayo_created_agents_test.go
    trigger_stats.go
    trigger_stats_test.go
    agent_capabilities.go
    agent_capabilities_test.go
```

### Test Helper

Create a shared test helper for in-memory SQLite:

```go
// internal/repositories/testutil_test.go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    
    // Run migrations
    err = migrations.Run(db)
    require.NoError(t, err)
    
    t.Cleanup(func() { db.Close() })
    return db
}
```

### Test Cases Per Repository

**user_sessions_test.go:**
- TestCreateSession - creates session with correct fields
- TestGetSession - retrieves existing session
- TestGetSession_NotFound - returns nil for missing session
- TestListSessions - pagination works correctly
- TestListSessions_FilterByAgent - filters by agent name
- TestDeleteSession - removes session and associated messages

**user_messages_test.go:**
- TestAddMessage - adds message to session
- TestGetMessages - retrieves messages in order
- TestGetMessages_Empty - returns empty slice
- TestGetMessages_Pagination - offset/limit work
- TestDeleteOldMessages - cleanup by date

**flow_runs_test.go:**
- TestRecordRun - stores run with all fields
- TestGetRun - retrieves by ID
- TestListRuns_ByFlow - filters by flow name
- TestListRuns_ByStatus - filters by success/failure
- TestListRuns_DateRange - filters by date
- TestGetRunStats - aggregates success rate

**ayo_created_agents_test.go:**
- TestRecordCreation - stores creation record
- TestGetCreatedAgents - lists agents created by @ayo
- TestRecordRefinement - updates refinement count
- TestGetRefinementHistory - shows refinement log
- TestMarkArchived - sets archived flag

**trigger_stats_test.go:**
- TestRecordTriggerRun - stores run
- TestGetRunCount - counts before permanent
- TestShouldBecomePermanent - threshold logic
- TestResetStats - clears stats after promotion

**agent_capabilities_test.go:**
- TestStoreCapabilities - stores with embeddings
- TestGetCapabilities - retrieves for agent
- TestSearchByEmbedding - vector similarity search
- TestInvalidateCache - clears on hash mismatch
- TestGetCapabilities_ExcludeUnrestricted - trust filtering

### Test Data

Create fixtures for realistic test data:

```go
// internal/repositories/fixtures_test.go
var testAgent = Agent{
    ID:   "test-agent-1",
    Name: "test-reviewer",
    TrustLevel: "sandboxed",
}

var testSession = Session{
    ID:        "session-1",
    AgentID:   "test-agent-1",
    CreatedAt: time.Now(),
}
```

## Acceptance Criteria

- [ ] Test file exists for each repository
- [ ] setupTestDB helper with in-memory SQLite
- [ ] All CRUD operations tested
- [ ] Edge cases covered (not found, empty, etc.)
- [ ] Pagination tests where applicable
- [ ] Filter tests where applicable
- [ ] All tests pass with `go test ./internal/repositories/...`
- [ ] Code coverage > 80% for repository package

