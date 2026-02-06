package sandbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/providers"
)

const (
	defaultToolTimeout  = 30 * time.Second
	outputLimitBytes    = 64 * 1024
)

// BashParams defines the parameters for the bash tool.
type BashParams struct {
	Command        string `json:"command" description:"Command to run (will be executed via /bin/sh -c)"`
	Description    string `json:"description" description:"Brief human-readable description of what this command does (e.g. 'Installing dependencies', 'Running tests')"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" description:"Optional timeout in seconds"`
	WorkingDir     string `json:"working_dir,omitempty" description:"Optional working directory scoped to the project root"`
}

// BashResult represents the result of a bash command execution.
type BashResult struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exit_code"`
	TimedOut  bool   `json:"timed_out,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
	Error     string `json:"error,omitempty"`
}

// String returns a formatted string representation of the result.
func (r BashResult) String() string {
	var parts []string

	if r.Stdout != "" {
		parts = append(parts, fmt.Sprintf("stdout:\n%s", r.Stdout))
	}
	if r.Stderr != "" {
		parts = append(parts, fmt.Sprintf("stderr:\n%s", r.Stderr))
	}
	if r.Error != "" {
		parts = append(parts, fmt.Sprintf("error: %s", r.Error))
	}
	if r.TimedOut {
		parts = append(parts, "timed out")
	}
	if r.Truncated {
		parts = append(parts, "(output truncated)")
	}

	parts = append(parts, fmt.Sprintf("exit_code: %d", r.ExitCode))

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "\n"
		}
		result += p
	}
	return result
}

// Executor executes bash commands in a sandbox.
type Executor struct {
	provider  providers.SandboxProvider
	sandboxID string
	baseDir   string
	user      string // User to run commands as (empty = root)
}

// NewExecutor creates a new sandbox executor.
func NewExecutor(provider providers.SandboxProvider, sandboxID, baseDir, user string) *Executor {
	return &Executor{
		provider:  provider,
		sandboxID: sandboxID,
		baseDir:   baseDir,
		user:      user,
	}
}

// Exec executes a command in the sandbox.
func (e *Executor) Exec(ctx context.Context, params BashParams) (BashResult, error) {
	if params.Command == "" {
		return BashResult{}, errors.New("command is required")
	}

	timeout := defaultToolTimeout
	if params.TimeoutSeconds > 0 {
		timeout = time.Duration(params.TimeoutSeconds) * time.Second
	}

	workingDir := e.baseDir
	if params.WorkingDir != "" {
		workingDir = params.WorkingDir
	}

	result, err := e.provider.Exec(ctx, e.sandboxID, providers.ExecOptions{
		Command:    params.Command,
		WorkingDir: workingDir,
		Timeout:    timeout,
		User:       e.user,
	})
	if err != nil {
		return BashResult{}, fmt.Errorf("sandbox exec: %w", err)
	}

	bashResult := BashResult{
		Stdout:    truncateOutput(result.Stdout, outputLimitBytes),
		Stderr:    truncateOutput(result.Stderr, outputLimitBytes),
		ExitCode:  result.ExitCode,
		TimedOut:  result.TimedOut,
		Truncated: result.Truncated || len(result.Stdout) > outputLimitBytes || len(result.Stderr) > outputLimitBytes,
	}

	if result.TimedOut {
		bashResult.Error = "command timed out"
	}

	return bashResult, nil
}

// NewBashTool creates a sandboxed bash tool for Fantasy.
func NewBashTool(executor *Executor) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"bash",
		"Execute a shell command in the sandbox and return stdout/stderr",
		func(ctx context.Context, params BashParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			result, err := executor.Exec(ctx, params)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

func truncateOutput(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
