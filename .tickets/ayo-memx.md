---
id: ayo-memx
status: closed
deps: [ayo-xfu3]
links: []
created: 2026-02-24T01:30:00Z
type: epic
priority: 0
assignee: Alex Cabrera
tags: [gtm, phase6, memory]
---
# Phase 6: Memory & Interactive Mode

Memory is the key differentiator that makes ayo grow and adapt with the user over time. This phase promotes memory to a first-class citizen and rewrites the interactive TUI.

## Goals

- Expose memory system through CLI commands
- Add memory tools for agents to store/search explicitly
- Implement squad-scoped memories for team knowledge
- Rewrite interactive mode with simpler, faster TUI
- Document memory system comprehensively

## Memory System Current State

The memory system already exists and works well:
- Async formation via small LLM
- Semantic deduplication (embedding similarity)
- Supersession chains preserving history
- SQLite + Zettelkasten dual storage
- Multi-scope: global, agent, path, hybrid

## What's Missing

1. **User visibility**: No way for users to see/manage memories via CLI
2. **Agent control**: Agents can't explicitly store important memories
3. **Squad scope**: Memories aren't shared across squad agents
4. **Documentation**: Memory system is undocumented

## Interactive Mode Problems

Current TUI is slow because:
- Full re-render on every 80ms tick
- Complex triple abstraction for tool rendering
- Message duplication tracking overhead
- Unnecessary viewport management

## Child Tickets

### Memory
- `ayo-mem1`: Add memory CLI commands
- `ayo-mem2`: Add memory tools for agents
- `ayo-mem3`: Implement squad-scoped memories
- `ayo-mem4`: Add memory export/import
- `ayo-zett`: **Zettelkasten tools and embedding-note linking**

### Interactive Mode
- `ayo-evnt`: **Event rendering architecture** (CLI-first, non-blocking)
- `ayo-glam`: **Glamour initialization optimization** (singleton pattern)
- `ayo-tui1`: Design simplified interactive mode
- `ayo-tui2`: Implement streaming text renderer
- `ayo-tui3`: Implement inline tool display
- `ayo-tui4`: Remove old TUI code

### Documentation
- `ayo-mem5`: Document memory system
