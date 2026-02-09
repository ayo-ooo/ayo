---
id: ase-muan
status: closed
deps: [ase-sjha]
links: []
created: 2026-02-09T03:29:11Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for migration system

## Background

The SQLite migration system (ase-sjha) manages database schema changes. The migration runner needs unit tests.

## Test Cases

### Migration Runner Tests

```go
func TestRunMigrations_FreshDB(t *testing.T) {
    db := openTestDB(t, ":memory:")
    
    err := RunMigrations(db)
    
    assert.NoError(t, err)
    
    // Verify migrations table exists
    var count int
    db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
    assert.Greater(t, count, 0)
}

func TestRunMigrations_Idempotent(t *testing.T) {
    db := openTestDB(t, ":memory:")
    
    err := RunMigrations(db)
    assert.NoError(t, err)
    
    // Run again - should be no-op
    err = RunMigrations(db)
    assert.NoError(t, err)
}

func TestRunMigrations_VersionTracking(t *testing.T) {
    db := openTestDB(t, ":memory:")
    
    RunMigrations(db)
    
    // Verify version recorded
    var version int
    db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
    assert.Equal(t, len(Migrations), version)
}

func TestRunMigrations_PartialUpgrade(t *testing.T) {
    db := openTestDB(t, ":memory:")
    
    // Simulate DB at version 2
    createMigrationsTable(db)
    insertMigration(db, 1, "Initial")
    insertMigration(db, 2, "Second")
    
    // Run all migrations (should only run 3+)
    err := RunMigrations(db)
    assert.NoError(t, err)
    
    // Verify new versions applied
    var version int
    db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
    assert.Equal(t, len(Migrations), version)
}

func TestRunMigrations_FailureRollback(t *testing.T) {
    db := openTestDB(t, ":memory:")
    
    // Add a migration that will fail
    badMigration := Migration{
        Version:     999,
        Description: "This will fail",
        Up: func(tx *sql.Tx) error {
            return errors.New("intentional failure")
        },
    }
    
    oldMigrations := Migrations
    Migrations = append(Migrations, badMigration)
    defer func() { Migrations = oldMigrations }()
    
    err := RunMigrations(db)
    assert.Error(t, err)
    
    // Verify bad migration not recorded
    var count int
    db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 999").Scan(&count)
    assert.Equal(t, 0, count)
}
```

### Individual Migration Tests

```go
func TestMigration_UserSessions(t *testing.T) {
    db := openTestDB(t, ":memory:")
    RunMigrations(db)
    
    // Verify table exists with correct schema
    _, err := db.Exec(`
        INSERT INTO user_sessions (id, agent_id, created_at)
        VALUES ('test', 'agent', datetime('now'))
    `)
    assert.NoError(t, err)
}

func TestMigration_FlowRuns(t *testing.T) {
    db := openTestDB(t, ":memory:")
    RunMigrations(db)
    
    _, err := db.Exec(`
        INSERT INTO flow_runs (id, flow_name, status, started_at)
        VALUES ('run1', 'test-flow', 'success', datetime('now'))
    `)
    assert.NoError(t, err)
}

func TestMigration_AyoCreatedAgents(t *testing.T) {
    db := openTestDB(t, ":memory:")
    RunMigrations(db)
    
    _, err := db.Exec(`
        INSERT INTO ayo_created_agents (agent_id, created_by, reason, created_at)
        VALUES ('agent1', '@ayo', 'User needed help', datetime('now'))
    `)
    assert.NoError(t, err)
}

func TestMigration_TriggerStats(t *testing.T) {
    db := openTestDB(t, ":memory:")
    RunMigrations(db)
    
    _, err := db.Exec(`
        INSERT INTO trigger_stats (trigger_id, run_count, last_run_at)
        VALUES ('trig1', 5, datetime('now'))
    `)
    assert.NoError(t, err)
}

func TestMigration_AgentCapabilities(t *testing.T) {
    db := openTestDB(t, ":memory:")
    RunMigrations(db)
    
    _, err := db.Exec(`
        INSERT INTO agent_capabilities (id, agent_id, name, description, confidence)
        VALUES ('cap1', 'agent1', 'code-review', 'Reviews code', 0.95)
    `)
    assert.NoError(t, err)
}
```

### GetCurrentVersion Tests

```go
func TestGetCurrentVersion_NoMigrations(t *testing.T) {
    db := openTestDB(t, ":memory:")
    createMigrationsTable(db)
    
    version, err := GetCurrentVersion(db)
    assert.NoError(t, err)
    assert.Equal(t, 0, version)
}

func TestGetCurrentVersion_WithMigrations(t *testing.T) {
    db := openTestDB(t, ":memory:")
    RunMigrations(db)
    
    version, err := GetCurrentVersion(db)
    assert.NoError(t, err)
    assert.Equal(t, len(Migrations), version)
}

func TestGetCurrentVersion_NoTable(t *testing.T) {
    db := openTestDB(t, ":memory:")
    
    version, err := GetCurrentVersion(db)
    assert.NoError(t, err)
    assert.Equal(t, 0, version)  // Treat as version 0
}
```

### Files to Create

1. `internal/db/migrations_test.go`

## Acceptance Criteria

- [ ] Fresh database migration test
- [ ] Idempotent migration test
- [ ] Version tracking test
- [ ] Partial upgrade test (skip already applied)
- [ ] Failure rollback test
- [ ] Individual table creation tests
- [ ] GetCurrentVersion tests
- [ ] All migrations compile and run
- [ ] All tests pass

