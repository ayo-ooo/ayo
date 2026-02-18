---
id: am-yu4b
status: open
deps: [am-q2ni]
links: []
created: 2026-02-18T03:15:33Z
type: task
priority: 2
assignee: Alex Cabrera
parent: am-8epk
---
# Add ayo squad schema commands

Add CLI commands for managing squad schemas.

## Context
- Users need to create and validate squad schemas
- Commands should be under 'ayo squad schema'

## Implementation
```bash
ayo squad schema init <squad>     # Create template input/output schemas
ayo squad schema validate <squad> # Validate schemas are valid JSON Schema
ayo squad schema show <squad>     # Display current schemas
```

## Files to Modify
- cmd/ayo/squad.go (add schema subcommand)

## Dependencies
- am-q2ni (schema loading)

## Acceptance
- ayo squad schema init creates template files
- ayo squad schema validate checks syntax
- ayo squad schema show displays schemas

