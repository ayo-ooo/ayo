---
id: am-3pxg
status: open
deps: [am-drxw, am-s0ya]
links: []
created: 2026-02-18T03:17:15Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-4i9o
---
# Implement lazy embedding invalidation

Check content hash before using embedding, re-embed if stale.

## Context
- On lookup, compare stored hash with computed hash
- If mismatch, re-embed and update index
- Slight latency on first access after change is acceptable

## Implementation
```go
// internal/capabilities/index.go

func (idx *EntityIndex) GetAgentEmbedding(ctx context.Context, ag agent.Agent) ([]float32, error) {
    currentHash := ComputeAgentHash(ag)
    
    stored, exists := idx.lookupAgent(ag.Handle)
    if exists && stored.ContentHash == currentHash {
        return stored.Embedding, nil // Cache hit
    }
    
    // Re-embed
    text := ag.Config.Description + "\n" + ag.System
    embedding, err := idx.embedder.Embed(ctx, text)
    if err != nil {
        return nil, err
    }
    
    // Update index
    idx.updateAgent(ag.Handle, currentHash, embedding)
    return embedding, nil
}
```

## Files to Modify
- internal/capabilities/index.go

## Dependencies
- am-drxw (EntityIndex struct)
- am-s0ya (hash computation)

## Acceptance
- Cache hit when hash matches
- Re-embed when hash differs
- Index updated after re-embedding

