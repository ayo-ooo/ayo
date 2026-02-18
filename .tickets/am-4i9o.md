---
id: am-4i9o
status: open
deps: []
links: []
created: 2026-02-18T03:12:49Z
type: epic
priority: 2
assignee: Alex Cabrera
---
# Embedding-Based Agent/Squad Selection

Build embedding-based selection for routing tasks to appropriate agents/squads. Use lazy hash invalidation for keeping embeddings in sync. Leverage existing CapabilitySearcher infrastructure.

## Acceptance Criteria

- Agent descriptions embedded and indexed
- Squad missions (from SQUAD.md) embedded and indexed
- Hash-based invalidation on read (content_hash stored with embedding)
- Semantic similarity ranking for dispatch decisions
- Integration with @ayo dispatch logic

