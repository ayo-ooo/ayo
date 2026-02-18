---
id: am-8epk
status: open
deps: []
links: []
created: 2026-02-18T03:12:39Z
type: epic
priority: 2
assignee: Alex Cabrera
---
# Squad I/O Schema System

Add optional input.jsonschema and output.jsonschema support to squads. Enable squads to act as structured functions (mixture of agents) with validated I/O. Include squad-lead routing logic.

## Acceptance Criteria

- Squads can have input.jsonschema and output.jsonschema
- Schema validation on dispatch and completion
- Squad-lead intercepts input by default
- Optional designated input agent via SQUAD.md
- Reject at boundary if schema validation fails
- Free-form fallback when no schemas defined

