---
id: ase-sjha
status: closed
deps: []
links: []
created: 2026-02-09T03:24:51Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-jep0
---
# Implement SQLite migration system

## Background

The agent orchestration system adds 5 new SQLite tables. We need a migration system to:
- Create tables on fresh install
- Upgrade existing databases when schema changes
- Track which migrations have run

## Why This Matters

Without migrations:
- Users upgrading would get "table not found" errors
- No way to evolve schema over time
- Database state becomes inconsistent

## Implementation Details

### Migration System

```go
// internal/db/migrations.go
type Migration struct {
    Version     int
    Description string
    Up          func(tx *sql.Tx) error
    Down        func(tx *sql.Tx) error  // Optional, for rollback
}

var Migrations = []Migration{
    {
        Version: 1,
        Description: "Initial schema",
        Up: func(tx *sql.Tx) error {
            // Existing tables...
        },
    },
    {
        Version: 2,
        Description: "Add user_sessions and user_messages tables",
        Up: func(tx *sql.Tx) error {
            _, err := tx.Exec(`
                CREATE TABLE user_sessions (...)
            `)
            return err
        },
    },
    // ... more migrations
}
```

### Migrations Table

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);
```

### Migration Runner

```go
func RunMigrations(db *sql.DB) error {
    // 1. Create migrations table if not exists
    // 2. Get current version
    // 3. Apply pending migrations in order
    // 4. Record each migration
}
```

### New Migrations Needed

| Version | Description |
|---------|-------------|
| N+1 | Add user_sessions and user_messages tables |
| N+2 | Add flow_runs table |
| N+3 | Add ayo_created_agents table |
| N+4 | Add trigger_stats table |
| N+5 | Add agent_capabilities table with embeddings |
| N+6 | Add trust_level column to agents table |

### Files to Create/Modify

1. `internal/db/migrations.go` - Migration system
2. `internal/db/migrations/` - Individual migration files
3. `internal/db/db.go` - Call RunMigrations on open

### CLI Commands

```bash
# Check migration status
ayo db status

# Run pending migrations (usually automatic)
ayo db migrate

# Show migration history
ayo db history
```

## Acceptance Criteria

- [ ] Migration system with version tracking
- [ ] schema_migrations table created automatically
- [ ] Migrations run on database open
- [ ] Each new table has its own migration
- [ ] ayo db status shows current version
- [ ] Migrations are idempotent (safe to run twice)
- [ ] Unit tests for migration runner

