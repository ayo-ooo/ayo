---
id: am-mh6x
status: closed
deps: []
links: []
created: 2026-02-18T03:18:55Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-okzf
---
# Add squad field to YAML flow step spec

Extend YAML flow step spec to support squad targeting.

## Context
- Flow steps can specify which squad to run in
- Enables cross-squad orchestration

## Implementation
```go
// internal/flows/spec.go

type Step struct {
    ID        string   `yaml:"id"`
    Type      string   `yaml:"type"`       // "shell" or "agent"
    Agent     string   `yaml:"agent"`      // @agent handle
    Squad     string   `yaml:"squad"`      // #squad handle (NEW)
    Prompt    string   `yaml:"prompt"`
    DependsOn []string `yaml:"depends_on"`
    // ...
}
```

Example usage:
```yaml
steps:
  - id: review
    type: agent
    agent: "@reviewer"
    squad: "#frontend-team"  # Run in this squad
    prompt: "Review the code"
```

## Files to Modify
- internal/flows/spec.go

## Acceptance
- Squad field parsed from YAML
- Validation: squad must exist
- Default: no squad (run in @ayo sandbox)

