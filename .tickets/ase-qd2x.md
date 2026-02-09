---
id: ase-qd2x
status: closed
deps: [ase-sjha]
links: []
created: 2026-02-09T03:04:23Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-jep0
---
# Add trigger_stats table

Create SQLite table to track trigger execution statistics and confidence levels.

## Background

Triggers are defined in flow YAML files, but their runtime stats live in SQLite:
- Track runs completed vs failed
- Calculate confidence for 'runs_before_permanent' feature
- Track last run time for debugging

## Schema

```sql
CREATE TABLE trigger_stats (
  trigger_id TEXT PRIMARY KEY,          -- Unique trigger identifier
  flow_name TEXT NOT NULL,              -- Which flow this trigger belongs to
  
  -- Execution stats
  runs_completed INTEGER DEFAULT 0,
  runs_failed INTEGER DEFAULT 0,
  last_run_at TIMESTAMP,
  last_error TEXT,                      -- Most recent error if any
  
  -- Confidence tracking
  runs_before_permanent INTEGER,        -- From flow spec, NULL = already permanent
  permanent BOOLEAN DEFAULT FALSE,      -- Graduated from provisional
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_trigger_stats_flow ON trigger_stats(flow_name);
```

## Implementation

1. Add migration file
2. Add Go types for TriggerStats
3. Add repository methods: GetOrCreateStats, RecordRun, RecordFailure, CheckPermanent
4. Integrate with trigger engine

## Files to modify

- internal/database/migrations/NNN_trigger_stats.sql (new)
- internal/database/models/trigger_stats.go (new)
- internal/database/repository.go (add methods)

## Acceptance Criteria

- Migration creates table successfully
- Stats increment atomically
- Permanent flag set when runs_completed >= runs_before_permanent
- Can query by flow name

