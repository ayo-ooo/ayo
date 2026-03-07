---
id: ayo-g02x
status: closed
deps: [ayo-nw5i]
links: []
created: 2026-03-07T21:13:38Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Implement LLM-based judging for evals

Create judge.go in internal/build/ that uses LLM to score actual output against expected output. Support configurable criteria and scoring models.

## Design

Implementation details:

1. Judge struct:
   - Uses provider/model from evals config
   - Runs comparison evaluation
   - Returns score and reasoning

2. Function signatures:
   - NewJudge(cfg build.Config) (*Judge, error)
   - Compare(ctx, input, expected, actual, criteria string) (JudgeResult, error)
   - Score(ctx, input, expected, actual, criteria string) (float64, string, error) - returns score, reasoning

3. JudgeResult struct:
   - Score (float64): 0-10 score
   - Reasoning (string): Explanation from LLM
   - Passed (bool): Whether it meets threshold (e.g., >= 7/10)

4. Prompt template for LLM:
   - System prompt: "You are an evaluator. Compare actual output to expected output based on criteria: {criteria}"
   - User prompt: "Input: {input}\n\nExpected: {expected}\n\nActual: {actual}\n\nRate accuracy 0-10 and explain."

Need to add to internal/build/judge.go

## Acceptance Criteria

1. Judge struct created with Compare and Score methods
2. Uses LLM to evaluate outputs
3. Returns structured score and reasoning
4. Handles JSON parsing of LLM responses

