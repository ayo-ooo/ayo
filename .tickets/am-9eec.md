---
id: am-9eec
status: open
deps: [am-gt91]
links: []
created: 2026-02-20T02:50:34Z
type: task
priority: 2
assignee: Alex Cabrera
tags: [squads, schemas]
---
# Validate squad input against input.jsonschema before dispatch

Squad I/O schemas (input.jsonschema, output.jsonschema) exist in squad directories but validation is not enforced during dispatch. The ValidateInput function should check against the schema if present.

## Design

In Squad.ValidateInput(), if input.jsonschema exists in the squad directory, parse the input as JSON and validate against the schema. Return validation errors to the user before routing.

## Acceptance Criteria

- 'ayo squad schema validate' command validates test input
- Dispatch with invalid JSON input returns schema validation error
- Dispatch with valid input proceeds normally

