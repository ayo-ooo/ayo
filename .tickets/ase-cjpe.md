---
id: ase-cjpe
status: closed
deps: []
links: []
created: 2026-02-09T03:03:36Z
type: epic
priority: 2
assignee: Alex Cabrera
parent: ase-zlew
---
# Capability Inference

Infer agent capabilities from system prompts, skills, and schemas. Store in SQLite with embeddings for semantic search. Enables @ayo to select appropriate agents for tasks.

## Acceptance Criteria

- Capabilities inferred via LLM analysis
- Stored in SQLite with vector embeddings
- Semantic search for capability matching
- Cache invalidation via definition hash
- ayo agents capabilities CLI

