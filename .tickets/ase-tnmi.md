---
id: ase-tnmi
status: closed
deps: [ase-sjha]
links: []
created: 2026-02-09T03:04:14Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-jep0
---
# Add ayo_created_agents table

Create SQLite table to track agents that @ayo creates dynamically. These are specialized agents created based on usage patterns.

## Background

@ayo can create specialized agents (e.g., @science-researcher) when it recognizes repeated patterns. These agents:
- Are tracked separately from user-created or plugin agents
- Have usage metrics for refinement decisions
- Can be archived (hidden but not deleted)
- May be promoted to user-owned agents

## Schema

```sql
CREATE TABLE ayo_created_agents (
  handle TEXT PRIMARY KEY,              -- '@science-researcher'
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  -- Learning metrics
  invocation_count INTEGER DEFAULT 0,
  success_count INTEGER DEFAULT 0,
  failure_count INTEGER DEFAULT 0,
  last_used_at TIMESTAMP,
  
  -- Evolution tracking
  system_prompt_version INTEGER DEFAULT 1,
  refinement_notes TEXT,                -- JSON array of refinement history
  
  -- Lifecycle
  confidence REAL DEFAULT 0.0,          -- 0.0-1.0, increases with success
  archived BOOLEAN DEFAULT FALSE,       -- Hidden but not deleted
  promoted_to TEXT                      -- New handle if promoted, NULL otherwise
);

CREATE INDEX idx_ayo_agents_archived ON ayo_created_agents(archived);
CREATE INDEX idx_ayo_agents_confidence ON ayo_created_agents(confidence);
```

## Implementation

1. Add migration file
2. Add Go types for AyoCreatedAgent
3. Add repository methods: CreateAgent, IncrementUsage, RecordSuccess/Failure, Archive, Promote, GetByConfidence
4. Add method to check if handle is @ayo-created

## Files to modify

- internal/database/migrations/NNN_ayo_created_agents.sql (new)
- internal/database/models/ayo_agent.go (new)
- internal/database/repository.go (add methods)

## Acceptance Criteria

- Migration creates table successfully
- Metrics increment atomically
- Confidence calculation works correctly
- Archive/promote update state properly

