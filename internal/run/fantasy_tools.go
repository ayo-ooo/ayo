package run

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/plugins"
)

// Tool parameter types for Fantasy

// BashParams defines the parameters for the bash tool.
type BashParams struct {
	Command        string `json:"command" description:"Command to run (will be executed via /bin/sh -c)"`
	Description    string `json:"description" description:"Brief human-readable description of what this command does (e.g. 'Installing dependencies', 'Running tests')"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" description:"Optional timeout in seconds"`
	WorkingDir     string `json:"working_dir,omitempty" description:"Optional working directory scoped to the project root"`
}

// AgentCallParams defines the parameters for the agent_call tool.
type AgentCallParams struct {
	Agent          string `json:"agent" description:"The agent handle to call (e.g., '@ayo'). Must be a builtin agent."`
	Prompt         string `json:"prompt" description:"The prompt/question to send to the agent"`
	Model          string `json:"model,omitempty" description:"Model to use for the sub-agent (e.g., 'claude-sonnet-4'). If not specified, uses the sub-agent's default."`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" description:"Optional timeout in seconds (default 120, max 300)"`
}

const (
	fantasyDefaultToolTimeout = 30 * time.Second
	fantasyOutputLimitBytes   = 64 * 1024
)

// NewBashTool creates the bash tool for Fantasy.
func NewBashTool(baseDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"bash",
		"Execute a shell command and return stdout/stderr",
		func(ctx context.Context, params BashParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if strings.TrimSpace(params.Command) == "" {
				return fantasy.NewTextErrorResponse("command is required; provide a string like {\"command\":\"echo hello world\"}"), nil
			}

			timeout := fantasyDefaultToolTimeout
			if params.TimeoutSeconds > 0 {
				timeout = time.Duration(params.TimeoutSeconds) * time.Second
			}

			execCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			stdoutBuf := &fantasyLimitedBuffer{max: fantasyOutputLimitBytes}
			stderrBuf := &fantasyLimitedBuffer{max: fantasyOutputLimitBytes}

			workingDir, err := fantasyResolveWorkingDir(baseDir, params.WorkingDir)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("invalid working_dir: %w", err)
			}

			cmd := exec.CommandContext(execCtx, "/bin/sh", "-c", params.Command)
			cmd.Stdout = stdoutBuf
			cmd.Stderr = stderrBuf
			cmd.Dir = workingDir

			runErr := cmd.Run()

			result := fantasyBashResult{
				Stdout:    stdoutBuf.String(),
				Stderr:    stderrBuf.String(),
				Truncated: stdoutBuf.truncated || stderrBuf.truncated,
			}

			if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
				result.TimedOut = true
				result.ExitCode = -1
				result.Error = "bash timed out"
				return fantasy.NewTextResponse(result.String()), nil
			}

			if runErr != nil {
				var exitErr *exec.ExitError
				if errors.As(runErr, &exitErr) {
					result.ExitCode = exitErr.ExitCode()
				} else {
					result.ExitCode = -1
				}
				result.Error = runErr.Error()
				return fantasy.NewTextResponse(result.String()), nil
			}

			if cmd.ProcessState != nil {
				result.ExitCode = cmd.ProcessState.ExitCode()
			}

			return fantasy.NewTextResponse(result.String()), nil
		},
	)
}

// fantasyBashResult mirrors toolResult for JSON serialization.
type fantasyBashResult struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exit_code"`
	TimedOut  bool   `json:"timed_out,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (r fantasyBashResult) String() string {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf(`{"error":"marshal error: %v"}`, err)
	}
	return string(data)
}

// FantasyToolSet wraps Fantasy tools for use with the runner.
type FantasyToolSet struct {
	tools       []fantasy.AgentTool
	allowedList []string
	baseDir     string
	depth       int
}

// NewFantasyToolSet creates a Fantasy tool set from allowed tool names.
func NewFantasyToolSet(allowed []string) FantasyToolSet {
	baseDir, _ := os.Getwd()
	return NewFantasyToolSetWithDepth(allowed, baseDir, 0)
}

// NewFantasyToolSetWithBaseDir creates a Fantasy tool set with explicit base directory.
func NewFantasyToolSetWithBaseDir(allowed []string, baseDir string) FantasyToolSet {
	return NewFantasyToolSetWithDepth(allowed, baseDir, 0)
}

// NewFantasyToolSetWithDepth creates a Fantasy tool set with explicit base directory and depth.
func NewFantasyToolSetWithDepth(allowed []string, baseDir string, depth int) FantasyToolSet {
	// Default to bash and plan if no tools specified
	if len(allowed) == 0 {
		allowed = []string{"bash", "plan"}
	}

	var tools []fantasy.AgentTool
	loadedTools := make(map[string]bool)

	for _, name := range allowed {
		switch name {
		case "bash":
			tools = append(tools, NewBashTool(baseDir))
			loadedTools[name] = true
		case "plan":
			tools = append(tools, NewPlanTool())
			loadedTools[name] = true
		// agent_call is added separately when needed
		default:
			// Try to load as external tool from plugins
			if tool := loadExternalTool(name, baseDir, depth); tool != nil {
				tools = append(tools, tool)
				loadedTools[name] = true
			}
		}
	}

	return FantasyToolSet{tools: tools, allowedList: allowed, baseDir: baseDir, depth: depth}
}

// Tools returns the Fantasy agent tools.
func (ts FantasyToolSet) Tools() []fantasy.AgentTool {
	return ts.tools
}

// HasTool checks if a tool name is in the allowed list.
func (ts FantasyToolSet) HasTool(name string) bool {
	for _, t := range ts.allowedList {
		if t == name {
			return true
		}
	}
	return false
}

// AddAgentCallTool adds the agent_call tool with the given executor.
func (ts *FantasyToolSet) AddAgentCallTool(executor func(ctx context.Context, params AgentCallParams, call fantasy.ToolCall) (fantasy.ToolResponse, error)) {
	ts.tools = append(ts.tools, fantasy.NewAgentTool(
		"agent_call",
		"Call a builtin agent as a subprocess and get its response. Use this to delegate specialized tasks to other agents.",
		executor,
	))
}

// resolveWorkingDir is shared between old and new tool implementations.
// Already defined in tools.go, but we need it here too for the Fantasy tools.
func fantasyResolveWorkingDir(baseDir, workingDirArg string) (string, error) {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(workingDirArg) == "" {
		return absBase, nil
	}

	// If workingDirArg is already an absolute path, use it directly
	// but still validate it's within baseDir
	var absTarget string
	if filepath.IsAbs(workingDirArg) {
		absTarget = workingDirArg
	} else {
		absTarget, err = filepath.Abs(filepath.Join(absBase, workingDirArg))
		if err != nil {
			return "", err
		}
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") || (rel == "." && absTarget != absBase) {
		return "", fmt.Errorf("working_dir must stay within %s", absBase)
	}
	if statErr := fantasyEnsureDir(absTarget); statErr != nil {
		return "", statErr
	}
	return absTarget, nil
}

func fantasyEnsureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("working_dir is not a directory: %s", path)
	}
	return nil
}

// limitedBuffer is already defined in tools.go, but Fantasy tools use their own
// to avoid import cycles if we split packages later.
type fantasyLimitedBuffer struct {
	buf       bytes.Buffer
	max       int
	truncated bool
}

func (l *fantasyLimitedBuffer) Write(p []byte) (int, error) {
	if l.max <= 0 {
		return len(p), nil
	}
	remaining := l.max - l.buf.Len()
	if remaining > 0 {
		if len(p) <= remaining {
			_, _ = l.buf.Write(p)
		} else {
			_, _ = l.buf.Write(p[:remaining])
		}
	}
	if len(p) > remaining && remaining >= 0 {
		l.truncated = true
	}
	return len(p), nil
}

func (l *fantasyLimitedBuffer) String() string {
	return l.buf.String()
}

// loadExternalTool attempts to load a tool from installed plugins.
// Returns nil if the tool is not found.
func loadExternalTool(toolName string, baseDir string, depth int) fantasy.AgentTool {
	// Load plugin registry
	registry, err := plugins.LoadRegistry()
	if err != nil {
		return nil
	}

	// Search all enabled plugins for this tool
	for _, plugin := range registry.ListEnabled() {
		for _, tool := range plugin.Tools {
			if tool == toolName {
				// Load tool definition
				def, err := plugins.LoadToolDefinition(plugin.Path, toolName)
				if err != nil {
					continue
				}

				return NewExternalTool(def, plugin.Path, baseDir, depth)
			}
		}
	}

	return nil
}
