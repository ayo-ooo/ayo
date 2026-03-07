package build

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexcabrera/ayo/internal/build/types"
)

// EvalCase represents a single evaluation test case
type EvalCase struct {
	Description string          // Human-readable description (optional)
	Input       map[string]any  // JSON input to agent
	Expected    map[string]any  // JSON expected output
	Criteria    string          // Override criteria for this case (optional)
}

// EvalResult represents the result of running an evaluation
type EvalResult struct {
	Case        EvalCase
	Actual      map[string]any
	ActualJSON  string
	Score       float64
	Reasoning   string
	Passed      bool
	Error       error
}

// ParseEvalsCSV parses an evals.csv file and returns test cases
// CSV format (header row required):
//   - description (optional): Human-readable test case name
//   - input (required): JSON input to agent
//   - expected (required): JSON expected output
//   - criteria (optional): Override default criteria for this case
func ParseEvalsCSV(path string) ([]EvalCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open evals.csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read evals.csv: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("evals.csv is empty")
	}

	// Get headers
	headers := records[0]
	if len(headers) < 3 {
		return nil, fmt.Errorf("evals.csv must have at least 3 columns: description, input, expected, [criteria]")
	}

	// Map column names to indices
	colIndex := map[string]int{}
	for i, h := range headers {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}

	inputIdx, ok := colIndex["input"]
	if !ok {
		return nil, fmt.Errorf("evals.csv missing required 'input' column")
	}

	expectedIdx, ok := colIndex["expected"]
	if !ok {
		return nil, fmt.Errorf("evals.csv missing required 'expected' column")
	}

	descriptionIdx, _ := colIndex["description"]
	criteriaIdx, _ := colIndex["criteria"]

	// Parse rows
	cases := make([]EvalCase, 0, len(records)-1)
	for _, row := range records[1:] {
		// Skip empty rows
		if len(row) == 0 || (len(row) == 1 && row[0] == "") {
			continue
		}

		if len(row) <= max(inputIdx, expectedIdx) {
			return nil, fmt.Errorf("row %d: insufficient columns", len(cases)+2)
		}

		// Parse input JSON
		inputJSON := strings.TrimSpace(row[inputIdx])
		var input map[string]any
		if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
			return nil, fmt.Errorf("row %d: invalid input JSON: %w", len(cases)+2, err)
		}

		// Parse expected JSON
		expectedJSON := strings.TrimSpace(row[expectedIdx])
		var expected map[string]any
		if err := json.Unmarshal([]byte(expectedJSON), &expected); err != nil {
			return nil, fmt.Errorf("row %d: invalid expected JSON: %w", len(cases)+2, err)
		}

		// Get optional fields
		description := ""
		if descriptionIdx >= 0 && len(row) > descriptionIdx {
			description = strings.TrimSpace(row[descriptionIdx])
		}

		criteria := ""
		if criteriaIdx >= 0 && len(row) > criteriaIdx {
			criteria = strings.TrimSpace(row[criteriaIdx])
		}

		cases = append(cases, EvalCase{
			Description: description,
			Input:       input,
			Expected:    expected,
			Criteria:    criteria,
		})
	}

	if len(cases) == 0 {
		return nil, fmt.Errorf("evals.csv contains no valid test cases")
	}

	return cases, nil
}

// EvalRunner handles running evaluations for test cases
type EvalRunner struct {
	config *types.Config
}

// NewEvalRunner creates a new evaluation runner
func NewEvalRunner(cfg *types.Config) *EvalRunner {
	return &EvalRunner{config: cfg}
}

// RunEval executes a single evaluation case
func (r *EvalRunner) RunEval(evalCase EvalCase) (map[string]any, string, error) {
	// For checkit, we'll execute the agent in-process
	// This is a placeholder - actual implementation will depend on how we want to run the agent
	// For now, we'll return a mock result

	// TODO: Implement actual agent execution
	// This could involve:
	// 1. Loading the agent from config
	// 2. Calling the agent with the input
	// 3. Getting the output as JSON

	// Placeholder: return empty output
	output := map[string]any{
		"message": "agent execution not yet implemented",
	}

	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	return output, string(outputJSON), nil
}

// RunAllEvals executes all evaluation cases
func (r *EvalRunner) RunAllEvals(evalPath string) ([]EvalResult, error) {
	// Resolve to absolute path
	if !filepath.IsAbs(evalPath) {
		evalPath = filepath.Join(".", evalPath)
	}

	// Parse evals CSV
	cases, err := ParseEvalsCSV(evalPath)
	if err != nil {
		return nil, fmt.Errorf("parse evals.csv: %w", err)
	}

	// Run each case
	results := make([]EvalResult, len(cases))
	for i, evalCase := range cases {
		actual, actualJSON, err := r.RunEval(evalCase)
		results[i] = EvalResult{
			Case:       evalCase,
			Actual:      actual,
			ActualJSON:  actualJSON,
			Error:       err,
		}
	}

	return results, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
