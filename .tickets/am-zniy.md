---
id: am-zniy
status: closed
deps: [am-mw6n]
links: []
created: 2026-02-18T03:17:40Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-4i9o
---
# Add ayo index CLI commands

Add CLI commands for managing the entity index.

## Context
- Users may want to inspect or rebuild index
- Useful for debugging selection issues

## Implementation
```bash
ayo index status    # Show index stats (count, last update)
ayo index rebuild   # Force re-embed all agents and squads
ayo index search "query"  # Test search without dispatch
```

## Files to Create
- cmd/ayo/index.go

## Dependencies
- am-mw6n (UnifiedSearcher)

## Acceptance
- ayo index status shows stats
- ayo index rebuild re-embeds everything
- ayo index search returns ranked results

