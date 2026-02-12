package sandbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/util"
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
	provider    providers.SandboxProvider
	sandboxID   string
	baseDir     string
	user        string // User to run commands as (empty = root)
	sessionID   string // Session ID for workspace creation
	agentHandle string // Agent handle for env vars
	workspaceDir string // Path to session workspace in sandbox
	env          map[string]string // Environment variables to inject
}

// NewExecutor creates a new sandbox executor.
func NewExecutor(provider providers.SandboxProvider, sandboxID, baseDir, user string) *Executor {
	return &Executor{
		provider:  provider,
		sandboxID: sandboxID,
		baseDir:   baseDir,
		user:      user,
		env:       make(map[string]string),
	}
}

// SetSession configures the executor with session and agent information.
// This sets up environment variables and prepares workspace creation.
func (e *Executor) SetSession(sessionID, agentHandle string) {
	e.sessionID = sessionID
	e.agentHandle = agentHandle
	if sessionID != "" {
		e.workspaceDir = fmt.Sprintf("/workspaces/%s", sessionID)
		e.env["WORKSPACE"] = e.workspaceDir
		e.env["SESSION_ID"] = sessionID
	}
	if agentHandle != "" {
		e.env["AGENT"] = agentHandle
	}
}

// CreateSessionWorkspace creates the session workspace directory in the sandbox.
// Call this after SetSession and before executing commands.
func (e *Executor) CreateSessionWorkspace(ctx context.Context) error {
	if e.workspaceDir == "" {
		return nil // No workspace configured
	}

	// Create workspace directory structure
	workspaceDirs := []string{
		e.workspaceDir,
		fmt.Sprintf("%s/mounted", e.workspaceDir),
		fmt.Sprintf("%s/scratch", e.workspaceDir),
		fmt.Sprintf("%s/shared", e.workspaceDir),
	}

	mkdirCmd := fmt.Sprintf("mkdir -p %s", joinPaths(workspaceDirs))
	_, err := e.provider.Exec(ctx, e.sandboxID, providers.ExecOptions{
		Command:    mkdirCmd,
		WorkingDir: "/",
		Timeout:    10 * time.Second,
		User:       "root", // Use root to create dirs, then chown
	})
	if err != nil {
		return fmt.Errorf("create workspace directories: %w", err)
	}

	// Chown to agent user if specified
	if e.user != "" && e.user != "root" {
		chownCmd := fmt.Sprintf("chown -R %s:%s %s", e.user, e.user, e.workspaceDir)
		_, err := e.provider.Exec(ctx, e.sandboxID, providers.ExecOptions{
			Command:    chownCmd,
			WorkingDir: "/",
			Timeout:    10 * time.Second,
			User:       "root",
		})
		if err != nil {
			return fmt.Errorf("chown workspace: %w", err)
		}
	}

	return nil
}

// WorkspaceDir returns the path to the session workspace in the sandbox.
func (e *Executor) WorkspaceDir() string {
	return e.workspaceDir
}

// joinPaths joins paths with spaces for shell command.
func joinPaths(paths []string) string {
	result := ""
	for i, p := range paths {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
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
		Env:        e.env,
	})
	if err != nil {
		return BashResult{}, fmt.Errorf("sandbox exec: %w", err)
	}

	bashResult := BashResult{
		Stdout:    util.TruncateRaw(result.Stdout, outputLimitBytes),
		Stderr:    util.TruncateRaw(result.Stderr, outputLimitBytes),
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


