---
id: ayo-899j
status: open
deps: [ayo-wt6w]
links: []
created: 2026-02-23T23:16:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-sqad
tags: [scheduler, cli]
---
# Add job monitoring interface

Add `ayo trigger status` command showing job status, history, and next run times. Use gocron event listeners to track job execution.

## Context

After job persistence is implemented (ayo-wt6w), users need visibility into:
- What jobs are registered
- When they last ran (and status)
- When they will run next
- History of recent runs

## CLI Commands

### List Active Triggers

```bash
ayo trigger list
# NAME              TYPE      SCHEDULE          NEXT RUN           LAST RUN
# health-check      interval  every 5m          2026-02-24 14:05   2026-02-24 14:00 ✓
# morning-report    daily     09:00             2026-02-25 09:00   2026-02-24 09:00 ✓
# weekly-summary    weekly    Mon,Fri 09:00     2026-02-28 09:00   2026-02-24 09:00 ✗
```

### Show Trigger Details

```bash
ayo trigger show health-check
# Name:           health-check
# Type:           interval
# Agent:          @monitor
# Schedule:       every 5m (singleton)
# Status:         active
# 
# Next Run:       2026-02-24 14:05:00 (in 3m)
# 
# Recent History:
# TIME                  DURATION    STATUS    OUTPUT
# 2026-02-24 14:00:00   12s         success   No issues found
# 2026-02-24 13:55:00   45s         success   No issues found
# 2026-02-24 13:50:00   8s          failed    Connection timeout
```

### View Run History

```bash
ayo trigger history health-check --limit 50
```

## Event Listeners

Use gocron's event listeners to track execution:

```go
// internal/daemon/trigger_engine.go
func (te *TriggerEngine) setupEventListeners() {
    te.scheduler.WithGlobalJobOptions(
        gocron.WithEventListeners(
            gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
                te.recordJobStart(jobID, jobName)
            }),
            gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
                te.recordJobSuccess(jobID, jobName)
            }),
            gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, err error) {
                te.recordJobFailure(jobID, jobName, err)
            }),
        ),
    )
}
```

## Database Schema

```sql
-- Extends job_runs table from ayo-wt6w
CREATE TABLE IF NOT EXISTS job_runs (
    id INTEGER PRIMARY KEY,
    job_name TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    status TEXT NOT NULL,  -- 'running', 'success', 'failed'
    error TEXT,
    output TEXT,
    FOREIGN KEY (job_name) REFERENCES jobs(name)
);

-- Keep last 100 runs per job
CREATE TRIGGER cleanup_old_runs AFTER INSERT ON job_runs
BEGIN
    DELETE FROM job_runs WHERE id IN (
        SELECT id FROM job_runs 
        WHERE job_name = NEW.job_name 
        ORDER BY started_at DESC 
        LIMIT -1 OFFSET 100
    );
END;
```

## Files to Create/Modify

1. **`internal/daemon/trigger_engine.go`** - Add event listeners
2. **`internal/daemon/db.go`** - Add job_runs table and queries
3. **`cmd/ayo/trigger_list.go`** - List command with status
4. **`cmd/ayo/trigger_show.go`** - Detailed trigger info
5. **`cmd/ayo/trigger_history.go`** (new) - History view

## Acceptance Criteria

- [ ] `trigger list` shows all triggers with next/last run
- [ ] `trigger show` displays detailed trigger info
- [ ] `trigger history` shows recent runs
- [ ] Event listeners capture job start/success/failure
- [ ] Run history is stored in SQLite
- [ ] Old runs are automatically cleaned up (keep 100)
- [ ] JSON output supported (`--json`)

## Testing

- Test event listeners capture all states
- Test history cleanup keeps only 100 runs
- Test CLI output formatting
- Test JSON output
- Test daemon restart preserves history
