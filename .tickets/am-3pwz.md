---
id: am-3pwz
status: closed
deps: []
links: []
created: 2026-02-18T03:19:29Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-1wsq
---
# Write unified architecture overview doc

Create comprehensive documentation explaining the unified architecture.

## Context
- Need clear mental model for Skills, Flows, Planners
- Explain relationships between primitives
- Include decision tree for choosing primitives

## Content
1. Core primitives: Agents, Squads, Skills, Flows, Planners
2. Relationship diagram
3. Decision tree:
   - 'Do I know the exact steps?' → Flow
   - 'Do I know the goal?' → Squad dispatch
   - 'Need specialized knowledge?' → Skill
   - 'Need to track work?' → Planner
4. Examples of each

## Files to Create
- docs/architecture.md

## Acceptance
- Clear definitions of all primitives
- Visual diagram (mermaid)
- Decision tree
- Cross-references to detailed docs

