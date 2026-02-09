---
id: ase-ncrx
status: closed
deps: [ase-sjha]
links: []
created: 2026-02-09T03:04:03Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-jep0
---
# Add flow_runs table for execution history

Create SQLite table to track flow execution history. Flow definitions live in YAML files, but execution stats live in the database.

## Background

Flow files should not contain runtime stats (they're version-controlled definitions). Execution history belongs in SQLite for:
- Querying last N runs of a flow
- Tracking success/failure rates
- Confidence calculations for triggers

## Schema

```sql
CREATE TABLE flow_runs (
  id TEXT PRIMARY KEY,
  flow_name TEXT NOT NULL,
  started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP,
  status TEXT NOT NULL CHECK (status IN ('running', 'success', 'failed')),
  trigger_id TEXT,  -- which trigger fired this, if any (NULL for manual)
  error TEXT,       -- error message if failed
  input_json TEXT,  -- input parameters
  output_json TEXT  -- output if successful
);

CREATE INDEX idx_flow_runs_name ON flow_runs(flow_name);
CREATE INDEX idx_flow_runs_status ON flow_runs(status);
CREATE INDEX idx_flow_runs_trigger ON flow_runs(trigger_id);
```

## Implementation

1. Add migration file
2. Add Go types for FlowRun
3. Add repository methods: StartRun, CompleteRun, FailRun, GetRecentRuns, GetRunsByFlow
4. Integrate with flow executor (to be implemented)

## Files to modify

- internal/database/migrations/NNN_flow_runs.sql (new)
- internal/database/models/flow.go (new)
- internal/database/repository.go (add methods)

## Acceptance Criteria

- Migration creates table successfully
- Repository methods handle all status transitions
- Can query runs by flow name or trigger
- Timestamps are set correctly

