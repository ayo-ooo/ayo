---
id: ayo-z9hi
status: completed
deps: [ayo-ujgk, ayo-z9oo]
links: ["epic:build-system-refactor", "blocks:gl1u,62q7"]
created: 2026-03-07T21:04:19Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Refactor squad system for team projects

Refactor internal/squads/ to work with team.toml projects instead of ~/.local/share/ayo/squads/

## Implementation Plan
1. Update squads package to work with team.toml projects
2. Remove old squad directory structure references
3. Implement team project loading and coordination
4. Update squad-related commands to use new structure
5. Ensure team builds work with multiple agents

