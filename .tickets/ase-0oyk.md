---
id: ase-0oyk
status: closed
deps: [ase-qsc7]
links: []
created: 2026-02-09T03:10:35Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-cjpe
---
# Implement semantic search for capabilities

## Background

Once capabilities are inferred and stored with vector embeddings, @ayo needs to search them semantically. When @ayo receives a task like "review this code for bugs", it should find agents with capabilities like "code-review", "bug-detection", "static-analysis" even if those exact terms weren't used.

## Why This Matters

Keyword matching fails because:
- User asks for "code review" but agent capability is "source code analysis"
- User wants "summarize this" but agent capability is "text condensation"

Semantic search using embeddings solves this by matching meaning, not words.

## Implementation Details

### Embedding Generation

Use the same LLM provider to generate embeddings for:
1. Capability descriptions (at storage time)
2. Search queries (at search time)

```go
// internal/capabilities/embeddings.go
type EmbeddingService struct {
    provider providers.Provider
}

func (s *EmbeddingService) Embed(text string) ([]float32, error) {
    // Use embedding model (e.g., text-embedding-3-small)
    return s.provider.Embed(text)
}
```

### Search Implementation

```go
// internal/capabilities/search.go
type CapabilitySearch struct {
    db       *sql.DB
    embedder *EmbeddingService
}

type SearchResult struct {
    AgentID     string
    AgentName   string
    Capability  Capability
    Similarity  float64  // Cosine similarity score
}

func (s *CapabilitySearch) Search(query string, limit int) ([]SearchResult, error) {
    // 1. Generate embedding for query
    queryEmbed, err := s.embedder.Embed(query)
    
    // 2. Query SQLite with vector similarity
    // Using sqlite-vec extension for vector operations
    rows, err := s.db.Query(`
        SELECT 
            c.agent_id,
            a.name as agent_name,
            c.name,
            c.description,
            c.confidence,
            vec_distance_cosine(c.embedding, ?) as distance
        FROM agent_capabilities c
        JOIN agents a ON c.agent_id = a.id
        WHERE a.trust_level != 'unrestricted'  -- Hide unrestricted agents
        ORDER BY distance ASC
        LIMIT ?
    `, queryEmbed, limit)
    
    // 3. Convert to SearchResults
}
```

### SQLite Vector Extension

This depends on sqlite-vec being available. If not available, fall back to:
1. Load all capabilities into memory
2. Compute cosine similarity in Go
3. Sort and return top matches

```go
func cosineSimilarity(a, b []float32) float64 {
    var dot, normA, normB float64
    for i := range a {
        dot += float64(a[i] * b[i])
        normA += float64(a[i] * a[i])
        normB += float64(b[i] * b[i])
    }
    return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

### Files to Create/Modify

1. `internal/capabilities/embeddings.go` - Embedding generation
2. `internal/capabilities/search.go` - Semantic search implementation
3. `internal/capabilities/similarity.go` - Cosine similarity for fallback
4. Modify `internal/capabilities/repository.go` - Add search methods

### Integration with @ayo

```go
// In @ayo's planning phase
func (ayo *Ayo) SelectAgentForTask(task string) (*Agent, error) {
    results, err := ayo.capSearch.Search(task, 5)
    if err != nil {
        return nil, err
    }
    
    if len(results) == 0 {
        return nil, ErrNoCapableAgent
    }
    
    // Return highest confidence match
    return ayo.agents.Get(results[0].AgentID)
}
```

## Acceptance Criteria

- [ ] EmbeddingService generates vectors for text
- [ ] CapabilitySearch queries by semantic similarity
- [ ] Fallback to in-memory search if sqlite-vec unavailable
- [ ] Unrestricted agents excluded from search results
- [ ] Results sorted by similarity score
- [ ] Unit tests with known similar/dissimilar queries
- [ ] Benchmark test for search performance

