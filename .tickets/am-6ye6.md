---
id: am-6ye6
status: closed
deps: [am-m5tt]
links: []
created: 2026-02-18T03:15:19Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-8epk
---
# Add input_accepts field to SQUAD.md frontmatter

Allow squads to designate a specific agent to accept input instead of squad lead.

## Context
- By default, squad lead (@ayo-in-squad) intercepts all input
- Squads can designate another agent via SQUAD.md frontmatter

## Implementation
Extend ConstitutionFrontmatter:
```yaml
---
name: frontend-team
input_accepts: "@reviewer"  # This agent receives input directly
---
```

```go
// internal/squads/context.go

type ConstitutionFrontmatter struct {
    Name         string         `yaml:"name"`
    Planners     PlannersConfig `yaml:"planners"`
    InputAccepts string         `yaml:"input_accepts"` // Agent handle
}
```

## Files to Modify
- internal/squads/context.go

## Dependencies
- am-m5tt (frontmatter parsing already added)

## Acceptance
- input_accepts field parsed from frontmatter
- Validated that agent exists in squad
- Defaults to squad lead if not specified

