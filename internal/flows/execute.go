package flows

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/oklog/ulid/v2"
)

// RunStatus represents the outcome of a flow run.
type RunStatus string

const (
	RunStatusRunning          RunStatus = "running"
	RunStatusSuccess          RunStatus = "success"
	RunStatusFailed           RunStatus = "failed"
	RunStatusError            RunStatus = "error"
	RunStatusTimeout          RunStatus = "timeout"
	RunStatusValidationFailed RunStatus = "validation_failed"
)

// RunOptions configures flow execution.
type RunOptions struct {
	Input      string            // JSON input (from arg)
	InputFile  string            // Path to input file
	Timeout    time.Duration     // Execution timeout (default 5 min)
	WorkingDir string            // Override working directory
	Validate   bool              // Validate only, don't run
	Env        map[string]string // Additional environment variables

	// History recording options
	History       *HistoryService // If set, records run history
	ParentRunID   string          // Parent run ID if this is a nested flow
	SessionID     string          // Session ID if triggered from a session
	AutoPrune     bool            // If true, prunes old runs after completion
	RetentionDays int             // Max age in days for pruning
	MaxRuns       int64           // Max runs to keep for pruning
}

// RunResult contains the outcome of a flow execution.
type RunResult struct {
	RunID     string
	Flow      *Flow
	Status    RunStatus
	ExitCode  int
	Stdout    string        // Captured stdout (should be JSON)
	Stderr    string        // Captured stderr (logs)
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	InputUsed string // Actual input JSON
	Error     error
}

// Run executes a flow and returns the result.
func Run(ctx context.Context, flow *Flow, opts RunOptions) (*RunResult, error) {
	result := &RunResult{
		RunID:     generateRunID(),
		Flow:      flow,
		StartTime: time.Now(),
	}

	// Resolve input
	input, err := resolveInput(opts)
	if err != nil {
		result.Status = RunStatusError
		result.Error = fmt.Errorf("resolve input: %w", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		recordHistoryIfEnabled(ctx, opts, flow, result, false)
		return result, nil
	}
	result.InputUsed = input

	// Validate input against schema
	inputValidated := flow.HasInputSchema()
	if err := ValidateInput(flow, input); err != nil {
		result.Status = RunStatusValidationFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		recordHistoryIfEnabled(ctx, opts, flow, result, inputValidated)
		return result, nil
	}

	// Validate only mode
	if opts.Validate {
		result.Status = RunStatusSuccess
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, nil
	}

	// Record start in history
	if opts.History != nil {
		runID, err := opts.History.RecordStart(ctx, flow, input, inputValidated, opts.ParentRunID, opts.SessionID)
		if err == nil {
			result.RunID = runID
		}
	}

	// Set timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(ctx, "bash", flow.Path, input)

	// Set working directory
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	} else {
		cmd.Dir = flow.Dir
	}

	// Set environment
	cmd.Env = buildEnv(flow, result.RunID, input, opts.Env)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	err = cmd.Run()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Handle result
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = RunStatusTimeout
			result.Error = fmt.Errorf("flow timed out after %v", timeout)
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Status = RunStatusFailed
			result.Error = fmt.Errorf("flow exited with code %d", result.ExitCode)
		} else {
			result.Status = RunStatusError
			result.Error = err
		}
	} else {
		result.Status = RunStatusSuccess
		result.ExitCode = 0
	}

	// Record completion in history
	recordHistoryIfEnabled(ctx, opts, flow, result, inputValidated)

	return result, nil
}

// RunStreaming executes a flow with real-time stderr streaming.
func RunStreaming(ctx context.Context, flow *Flow, opts RunOptions, stderrWriter io.Writer) (*RunResult, error) {
	result := &RunResult{
		RunID:     generateRunID(),
		Flow:      flow,
		StartTime: time.Now(),
	}

	// Resolve input
	input, err := resolveInput(opts)
	if err != nil {
		result.Status = RunStatusError
		result.Error = fmt.Errorf("resolve input: %w", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		recordHistoryIfEnabled(ctx, opts, flow, result, false)
		return result, nil
	}
	result.InputUsed = input

	// Validate input against schema
	inputValidated := flow.HasInputSchema()
	if err := ValidateInput(flow, input); err != nil {
		result.Status = RunStatusValidationFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		recordHistoryIfEnabled(ctx, opts, flow, result, inputValidated)
		return result, nil
	}

	// Validate only mode
	if opts.Validate {
		result.Status = RunStatusSuccess
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, nil
	}

	// Record start in history
	if opts.History != nil {
		runID, err := opts.History.RecordStart(ctx, flow, input, inputValidated, opts.ParentRunID, opts.SessionID)
		if err == nil {
			result.RunID = runID
		}
	}

	// Set timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(ctx, "bash", flow.Path, input)

	// Set working directory
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	} else {
		cmd.Dir = flow.Dir
	}

	// Set environment
	cmd.Env = buildEnv(flow, result.RunID, input, opts.Env)

	// Capture stdout, stream stderr
	var stdout bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdout

	// Tee stderr to both buffer and writer
	if stderrWriter != nil {
		cmd.Stderr = io.MultiWriter(&stderrBuf, stderrWriter)
	} else {
		cmd.Stderr = &stderrBuf
	}

	// Execute
	err = cmd.Run()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Stdout = stdout.String()
	result.Stderr = stderrBuf.String()

	// Handle result
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = RunStatusTimeout
			result.Error = fmt.Errorf("flow timed out after %v", timeout)
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Status = RunStatusFailed
			result.Error = fmt.Errorf("flow exited with code %d", result.ExitCode)
		} else {
			result.Status = RunStatusError
			result.Error = err
		}
	} else {
		result.Status = RunStatusSuccess
		result.ExitCode = 0
	}

	// Record completion in history
	recordHistoryIfEnabled(ctx, opts, flow, result, inputValidated)

	return result, nil
}

// resolveInput determines the input JSON from options.
func resolveInput(opts RunOptions) (string, error) {
	// 1. Explicit input argument
	if opts.Input != "" {
		return opts.Input, nil
	}

	// 2. Input file
	if opts.InputFile != "" {
		data, err := os.ReadFile(opts.InputFile)
		if err != nil {
			return "", fmt.Errorf("read input file: %w", err)
		}
		return string(data), nil
	}

	// 3. Default to empty object
	return "{}", nil
}

// buildEnv creates the environment for flow execution.
func buildEnv(flow *Flow, runID, input string, extra map[string]string) []string {
	// Start with current environment
	env := os.Environ()

	// Add flow-specific variables
	env = append(env,
		fmt.Sprintf("AYO_FLOW_NAME=%s", flow.Name),
		fmt.Sprintf("AYO_FLOW_RUN_ID=%s", runID),
		fmt.Sprintf("AYO_FLOW_DIR=%s", flow.Dir),
	)

	// Create temp input file for large inputs
	if len(input) > 1024 {
		tmpFile, err := os.CreateTemp("", "ayo-flow-input-*.json")
		if err == nil {
			tmpFile.WriteString(input)
			tmpFile.Close()
			env = append(env, fmt.Sprintf("AYO_FLOW_INPUT_FILE=%s", tmpFile.Name()))
		}
	}

	// Add extra environment variables
	for k, v := range extra {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

// generateRunID creates a unique run identifier.
func generateRunID() string {
	return ulid.Make().String()
}

// CleanupTempFiles removes any temp files created during execution.
func CleanupTempFiles(result *RunResult) {
	// Find and remove temp input files
	pattern := filepath.Join(os.TempDir(), "ayo-flow-input-*.json")
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		os.Remove(match)
	}
}

// recordHistoryIfEnabled records the run result in history if a HistoryService is configured.
func recordHistoryIfEnabled(ctx context.Context, opts RunOptions, flow *Flow, result *RunResult, inputValidated bool) {
	if opts.History == nil {
		return
	}

	// Map run status for history
	historyStatus := result.Status
	if historyStatus == RunStatusError {
		historyStatus = RunStatusFailed
	}

	// Determine error message
	var errorMsg string
	if result.Error != nil {
		errorMsg = result.Error.Error()
	}

	// Complete the run record
	_, _ = opts.History.RecordComplete(ctx, result.RunID, CompleteResult{
		Status:          historyStatus,
		ExitCode:        result.ExitCode,
		ErrorMessage:    errorMsg,
		OutputJSON:      result.Stdout,
		StderrLog:       result.Stderr,
		OutputValidated: flow.HasOutputSchema() && result.Status == RunStatusSuccess,
	}, result.StartTime)

	// Auto-prune if enabled
	if opts.AutoPrune && (opts.RetentionDays > 0 || opts.MaxRuns > 0) {
		_ = opts.History.Prune(ctx, opts.RetentionDays, opts.MaxRuns)
	}
}
