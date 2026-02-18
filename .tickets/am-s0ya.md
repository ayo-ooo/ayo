---
id: am-s0ya
status: closed
deps: []
links: []
created: 2026-02-18T03:17:08Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-4i9o
---
# Implement content hash computation

Compute content hashes for agents and squads to detect changes.

## Context
- Hash agent config.json + system.md content
- Hash squad SQUAD.md content
- Used for lazy invalidation of embeddings

## Implementation
```go
// internal/capabilities/hash.go (new file)

func ComputeAgentHash(ag agent.Agent) string {
    h := sha256.New()
    h.Write([]byte(ag.Config.Description))
    h.Write([]byte(ag.System))
    // Include skill names for completeness
    for _, skill := range ag.Skills {
        h.Write([]byte(skill.Name))
    }
    return hex.EncodeToString(h.Sum(nil))
}

func ComputeSquadHash(constitution *squads.Constitution) string {
    h := sha256.New()
    h.Write([]byte(constitution.Raw))
    return hex.EncodeToString(h.Sum(nil))
}
```

## Files to Create
- internal/capabilities/hash.go

## Acceptance
- Consistent hash for same content
- Different hash when content changes
- Fast computation

