---
id: am-mw6n
status: closed
deps: [am-3pxg]
links: []
created: 2026-02-18T03:17:27Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-4i9o
---
# Create UnifiedSearcher for agents and squads

Create a searcher that finds best match across both agents and squads.

## Context
- Extend existing CapabilitySearcher concept
- Return ranked results with type (agent or squad)
- Used by @ayo for dispatch decisions

## Implementation
```go
// internal/capabilities/unified_search.go (new file)

type UnifiedSearcher struct {
    index    *EntityIndex
    embedder embedding.Embedder
}

type SearchResult struct {
    Type       string  // "agent" or "squad"
    Handle     string  // @agent or #squad
    Score      float64 // Cosine similarity
    HasSchema  bool
}

func (s *UnifiedSearcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
    queryEmb, err := s.embedder.Embed(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // Score all agents and squads
    // Return top N by similarity
}

func (s *UnifiedSearcher) FindBest(ctx context.Context, query string) (*SearchResult, error) {
    results, err := s.Search(ctx, query, 1)
    if len(results) == 0 {
        return nil, nil // No match, @ayo should handle
    }
    return &results[0], nil
}
```

## Files to Create
- internal/capabilities/unified_search.go

## Dependencies
- am-3pxg (lazy invalidation)

## Acceptance
- Searches both agents and squads
- Returns ranked results
- Handles empty results gracefully

