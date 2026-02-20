---
id: am-yb26
status: open
deps: [am-gt91]
links: []
created: 2026-02-20T02:50:40Z
type: task
priority: 2
assignee: Alex Cabrera
tags: [squads, schemas]
---
# Validate squad output against output.jsonschema after agent response

After an agent completes work in a squad, the output should be validated against output.jsonschema if present. This ensures squad responses conform to expected formats.

## Design

In squad dispatch, after receiving agent response, if output.jsonschema exists, attempt to parse response as JSON and validate. Log warning if validation fails but still return response.

## Acceptance Criteria

- Agent output is validated against output.jsonschema
- Validation failures are logged as warnings
- Valid output is returned to caller normally

