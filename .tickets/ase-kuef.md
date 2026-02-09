---
id: ase-kuef
status: closed
deps: [ase-sjha]
links: []
created: 2026-02-09T03:04:35Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-jep0
---
# Add agent_capabilities table with embeddings

Create SQLite table to store inferred agent capabilities with vector embeddings for semantic search.

## Background

Agent capabilities are inferred by LLM analysis of system prompts, skills, and schemas. Stored in SQLite with embeddings because:
- Inference is expensive (LLM calls), so cache the results
- Need semantic search ('find agents that can translate')
- Cache invalidation via hash of agent definition

## Schema

```sql
CREATE TABLE agent_capabilities (
  handle TEXT PRIMARY KEY,
  definition_hash TEXT NOT NULL,        -- SHA256 of agent.json + system prompt
  
  -- Structured data
  capabilities_json TEXT NOT NULL,      -- {capabilities: [], input_types: [], output_types: [], chains_well_with: []}
  
  -- Embedding for semantic search
  capabilities_text TEXT NOT NULL,      -- Flattened text description
  capabilities_embedding BLOB,          -- Vector embedding (float32 array)
  
  inferred_at TIMESTAMP NOT NULL
);
```

## Implementation

1. Add migration file
2. Add Go types for AgentCapabilities
3. Add repository methods: UpsertCapabilities, GetByHandle, GetAll, InvalidateIfChanged
4. Add vector similarity search function (compute in Go, not SQL)
5. Consider sqlite-vss extension for future, but start with in-memory similarity

## Vector search approach

Store embeddings as BLOB, load into memory for similarity search:
```go
func (r *Repository) FindSimilarAgents(queryEmbedding []float32, limit int) ([]AgentCapabilities, error) {
    all, _ := r.GetAllCapabilities()
    // Compute cosine similarity in Go
    // Return top N
}
```

## Files to modify

- internal/database/migrations/NNN_agent_capabilities.sql (new)
- internal/database/models/capabilities.go (new)
- internal/database/repository.go (add methods)
- internal/database/vector.go (new - similarity functions)

## Acceptance Criteria

- Migration creates table successfully
- Can store and retrieve embeddings
- Definition hash invalidation works
- Similarity search returns relevant results

