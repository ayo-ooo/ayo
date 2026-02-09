---
id: ase-jep0
status: closed
deps: []
links: []
created: 2026-02-09T03:03:42Z
type: epic
priority: 2
assignee: Alex Cabrera
parent: ase-zlew
---
# SQLite Schema Extensions

Extend SQLite schema to support orchestration features: user sessions, flow runs, agent creation tracking, trigger stats, and capability cache.

## Acceptance Criteria

- user_sessions and user_messages tables
- flow_runs table for execution history  
- ayo_created_agents table for agent lifecycle
- trigger_stats table for confidence tracking
- agent_capabilities table with embeddings
- Migrations for all new tables

