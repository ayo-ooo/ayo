---
id: ayo-wt6w
status: open
deps: [ayo-q841]
links: []
created: 2026-02-23T23:15:43Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler, daemon, sqlite]
---
# Add job persistence in SQLite

Store scheduled jobs in SQLite for persistence across daemon restarts.

## Database Location

`~/.local/share/ayo/jobs.db`

## Schema

```sql
CREATE TABLE scheduled_jobs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,  -- 'cron', 'daily', 'weekly', 'monthly', 'once', 'interval'
    schedule TEXT NOT NULL,  -- JSON config
    agent TEXT NOT NULL,
    prompt TEXT,
    output_path TEXT,
    singleton BOOLEAN DEFAULT false,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE job_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    status TEXT NOT NULL,  -- 'running', 'success', 'failed', 'cancelled'
    error_message TEXT,
    output_location TEXT,
    FOREIGN KEY (job_id) REFERENCES scheduled_jobs(id)
);

CREATE INDEX idx_job_runs_job_id ON job_runs(job_id);
CREATE INDEX idx_job_runs_started_at ON job_runs(started_at);
```

## Schedule JSON Examples

```json
// Cron
{"cron": "0 9 * * *"}

// Daily
{"times": ["09:00", "17:00"], "days": ["monday", "tuesday", "wednesday", "thursday", "friday"]}

// Weekly
{"day": "monday", "time": "09:00"}

// Monthly
{"day": 1, "time": "09:00"}

// Once
{"at": "2026-02-24T14:00:00Z"}

// Interval
{"every": "5m"}
```

## Implementation

### Location
`internal/daemon/job_store.go`

### Interface

```go
type JobStore interface {
    // CRUD
    Create(job *ScheduledJob) error
    Get(id string) (*ScheduledJob, error)
    List() ([]*ScheduledJob, error)
    Update(job *ScheduledJob) error
    Delete(id string) error
    
    // Run history
    RecordRun(run *JobRun) error
    GetRecentRuns(jobID string, limit int) ([]*JobRun, error)
    
    // Lifecycle
    LoadAllEnabled() ([]*ScheduledJob, error)
}
```

### Daemon Integration

1. On daemon start: `LoadAllEnabled()` → register with gocron
2. On job CRUD via RPC: update DB + update gocron
3. On job execution: `RecordRun()` with status
4. Keep last 100 runs per job, prune older

## Files to Create/Modify

- Create `internal/daemon/job_store.go`
- Create `internal/daemon/job_store_test.go`
- Update `internal/daemon/trigger_engine.go` to use store
- Update daemon startup to restore jobs

## Testing

- Test CRUD operations
- Test daemon restart restores jobs
- Test run history recording
- Test concurrent job execution
