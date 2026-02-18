---
id: am-x8nc
status: open
deps: []
links: []
created: 2026-02-18T03:12:34Z
type: epic
priority: 2
assignee: Alex Cabrera
---
# Planner Plugin System

Extract tickets into a pluggable planner system with near-term and long-term slots. Planners are instantiated per-sandbox with sandbox-scoped state. Default implementations (ayo-todos, ayo-tickets) ship as plugins.

## Acceptance Criteria

- Planner plugin interface defined
- Near-term and long-term plugin slots
- Per-sandbox instantiation with context
- Default plugins extracted from current code
- Global defaults configurable via config.toml
- Squad-level override via SQUAD.md frontmatter

