---
id: am-drxw
status: closed
deps: []
links: []
created: 2026-02-18T03:17:02Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-4i9o
---
# Define EntityIndex struct for agent/squad embeddings

Create data structure for storing agent and squad embeddings with content hashes.

## Context
- Need to store embeddings for agents and squads
- Include content hash for lazy invalidation
- Build on existing capabilities infrastructure

## Implementation
```go
// internal/capabilities/index.go (new file or extend existing)

type EntityIndex struct {
    Agents []IndexedAgent
    Squads []IndexedSquad
    db     *sql.DB
}

type IndexedAgent struct {
    Handle      string
    Description string
    ContentHash string    // SHA256 of description + system.md
    Embedding   []float32
    HasSchema   bool
    UpdatedAt   time.Time
}

type IndexedSquad struct {
    Name        string
    Mission     string    // From SQUAD.md
    ContentHash string    // SHA256 of SQUAD.md
    Embedding   []float32
    HasSchema   bool
    UpdatedAt   time.Time
}
```

## Files to Create
- internal/capabilities/index.go

## Acceptance
- Structs defined with all fields
- SQLite storage for persistence
- Content hash stored alongside embedding

