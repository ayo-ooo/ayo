---
id: am-m5tt
status: closed
deps: [am-1gwl]
links: []
created: 2026-02-18T03:13:59Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Add planner override to SQUAD.md frontmatter

Allow squads to override default planners via SQUAD.md frontmatter.

## Context
- SQUAD.md parsing: internal/squads/context.go
- Squads can use different planners than the global default

## Implementation
Update Constitution parsing to extract frontmatter:
```yaml
---
name: frontend-team
planners:
  near_term: ayo-todos
  long_term: ayo-kanban  # Override default
---
# Mission
...
```

```go
type ConstitutionFrontmatter struct {
    Name     string         `yaml:"name"`
    Planners PlannersConfig `yaml:"planners"`
    Lead     string         `yaml:"lead"`  // For future use
}
```

## Files to Modify
- internal/squads/context.go (add frontmatter parsing)

## Dependencies
- am-1gwl (PlannersConfig type defined)

## Acceptance
- SQUAD.md frontmatter parsed
- Planners config extracted
- Falls back to global config if not specified

