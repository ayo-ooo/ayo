---
id: ayo-mem3
status: open
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ayo-memx
tags: [memory, squads]
---
# Implement squad-scoped memories

Allow memories to be shared across all agents in a squad.

## Design

### New Scope: `squad`

```go
const (
    ScopeGlobal = "global"
    ScopeAgent  = "agent"
    ScopePath   = "path"
    ScopeSquad  = "squad"  // NEW
)
```

### Squad Memory Storage

Squad memories are stored in the squad's sandbox:
- Path: `/.memories/` in squad sandbox
- Accessible to all squad agents
- Synced via host daemon

### Memory Formation in Squads

When an agent in a squad forms a memory:
1. Check if memory is squad-relevant (mentions other agents, shared workspace, coordination)
2. If yes, store with scope=squad
3. If no, store with scope=agent

### Memory Injection for Squad Agents

When building context for a squad agent:
1. Load global memories (standard)
2. Load agent-specific memories (standard)
3. Load squad memories (NEW)
4. Combine and deduplicate

## Configuration

```json
// ayo.json for squad
{
  "squad": {
    "memory": {
      "shared": true,           // Enable squad memory
      "inherit_global": true    // Also include global memories
    }
  }
}
```

## Implementation

### Files to Modify

- `internal/memory/memory.go` - Add squad scope support
- `internal/memory/formation.go` - Squad-aware formation
- `internal/agent/memory_context.go` - Load squad memories
- `internal/squads/service.go` - Initialize squad memory store

## Testing

- Test squad memory creation
- Test squad memory visible to all squad agents
- Test squad memory isolated from other squads
- Test memory formation categorizes correctly
