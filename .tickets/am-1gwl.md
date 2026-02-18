---
id: am-1gwl
status: closed
deps: []
links: []
created: 2026-02-18T03:13:53Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-x8nc
---
# Add planner configuration to config.toml

Add global planner defaults to the ayo configuration system.

## Context
- Config location: ~/.config/ayo/config.toml
- Config struct: internal/config/config.go

## Implementation
Add to Config struct:
```go
type PlannersConfig struct {
    NearTerm string `toml:"near_term"` // default: "ayo-todos"
    LongTerm string `toml:"long_term"` // default: "ayo-tickets"
}
```

Add to config.toml template:
```toml
[planners]
near_term = "ayo-todos"
long_term = "ayo-tickets"
```

## Files to Modify
- internal/config/config.go (add PlannersConfig struct and field)
- internal/config/defaults.go (add defaults)

## Acceptance
- PlannersConfig parsed from config.toml
- Defaults to ayo-todos and ayo-tickets
- Config accessible via config.Config.Planners

