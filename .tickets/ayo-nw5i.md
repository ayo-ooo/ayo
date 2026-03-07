---
id: ayo-nw5i
status: closed
deps: []
links: []
created: 2026-03-07T21:13:10Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Add evals configuration to config schema

Update config.toml schema to support evals configuration. Add [evals] section with fields for model, provider, and evaluation criteria.

## Design

Configuration format:

[evals]
enabled = true
file = "evals.csv"
judge_model = "claude-3-5-sonnet"
judge_provider = "anthropic"
criteria = "accuracy,helpfulness"

This allows users to specify:
- evals.csv location
- Which model/provider to use for judging
- Evaluation criteria (can be customized per use case)

Need to update internal/build/types/config.go to add EvalsConfig struct.

## Acceptance Criteria

1. EvalsConfig struct added to config.go
2. Config parsing supports [evals] section
3. Default values provided (empty/criteria-based)
4. Validation ensures required fields are present if evals enabled

