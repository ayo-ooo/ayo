---
id: ayo-uy1g
status: closed
deps: [ayo-nw5i]
links: []
created: 2026-03-07T21:13:24Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Implement eval CSV parser and runner

Create internal/build/eval.go for parsing evals.csv and running evaluations. Parse CSV with input/output pairs, execute agent for each input, and collect results.

Implementation details:

1. CSV format (header row required):
   - description (optional): Human-readable test case name
   - input (required): JSON input to agent
   - expected (required): JSON expected output
   - criteria (optional): Override default criteria for this case

2. EvalRunner struct:
   - Parse evals.csv
   - Run each test case
   - Collect actual outputs
   - Pass to judge for scoring

3. Function signatures:
   - ParseEvalsCSV(path string) ([]EvalCase, error)
   - RunEval(config build.Config, evalCase EvalCase) (string, error) - returns actual output
   - RunAllEvals(config build.Config, evalsPath string) ([]EvalResult, error)

4. Error handling:
   - Invalid JSON in input/expected columns
   - Missing required columns
   - Agent execution failures

Need to add to internal/build/eval.go

## Acceptance Criteria

1. ParseEvalsCSV correctly parses CSV files
2. EvalRunner struct created with RunAllEvals method
3. Handles CSV parsing errors gracefully
4. Returns structured EvalResult for each case

