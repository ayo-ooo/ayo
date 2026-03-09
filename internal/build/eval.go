package build

import (
	"encoding/json"
	"fmt"
	"os"
)

// SimpleEval represents a basic evaluation test case
type SimpleEval struct {
	Name     string          // Test case name
	Input    map[string]any  // JSON input to agent
	Expected map[string]any   // JSON expected output
}

// ParseSimpleEvals parses a simple JSON-based evaluation file
// Format: array of {name: string, input: object, expected: object}
func ParseSimpleEvals(path string) ([]SimpleEval, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read evals file: %w", err)
	}

	var evals []SimpleEval
	if err := json.Unmarshal(data, &evals); err != nil {
		return nil, fmt.Errorf("parse evals JSON: %w", err)
	}

	return evals, nil
}

// SimpleEvalResult represents the result of a simple evaluation
type SimpleEvalResult struct {
	Eval      SimpleEval
	Actual    map[string]any
	Passed    bool
	Error     error
}

// RunSimpleEval runs a single evaluation test
func RunSimpleEval(eval SimpleEval, actual map[string]any) SimpleEvalResult {
	// Simple comparison: check if actual matches expected
	passed := deepEqual(eval.Expected, actual)
	
	return SimpleEvalResult{
		Eval:      eval,
		Actual:    actual,
		Passed:    passed,
		Error:     nil,
	}
}

// deepEqual compares two maps recursively
func deepEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, aVal := range a {
		bVal, exists := b[key]
		if !exists {
			return false
		}

		// Simple type comparison - for more complex cases, use reflection
		// This is sufficient for basic JSON comparison
		aJSON, _ := json.Marshal(aVal)
		bJSON, _ := json.Marshal(bVal)
		if string(aJSON) != string(bJSON) {
			return false
		}
	}

	return true
}