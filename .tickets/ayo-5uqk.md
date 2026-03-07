---
id: ayo-5uqk
status: closed
deps: [ayo-nw5i, ayo-uy1g, ayo-g02x]
links: []
created: 2026-03-07T21:13:54Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Integrate evals into checkit command

Update cmd/ayo/checkit.go to run evals if evals enabled and evals CSV exists. Display scoring results with pass/fail summary.

## Design

Implementation: Update runCheckit to check evals.enabled, parse CSV, run evals, judge results, display summary. Support --evals-threshold flag (default 7.0) and --evals-only flag. Exit code 0 on success, 1 on failure, 2 on errors.

## Acceptance Criteria

1. checkit runs evals when enabled
2. Displays scoring results for each test
3. Shows summary with pass/fail counts
4. Respects threshold flag
5. Exits with correct codes

