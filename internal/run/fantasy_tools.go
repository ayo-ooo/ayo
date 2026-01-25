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

	"github.com/alexcabrera/ayo/internal/memory"
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
	memoryQueue *memory.Queue // Optional async memory queue
}

// NewFantasyToolSet creates a Fantasy tool set from allowed tool names.
func NewFantasyToolSet(allowed []string) FantasyToolSet {
	baseDir, _ := os.Getwd()
	return NewFantasyToolSetWithOptions(allowed, baseDir, nil)
}

// NewFantasyToolSetWithBaseDir creates a Fantasy tool set with explicit base directory.
func NewFantasyToolSetWithBaseDir(allowed []string, baseDir string) FantasyToolSet {
	return NewFantasyToolSetWithOptions(allowed, baseDir, nil)
}

// NewFantasyToolSetWithOptions creates a Fantasy tool set with all options.
func NewFantasyToolSetWithOptions(allowed []string, baseDir string, memQueue *memory.Queue) FantasyToolSet {
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}

	// Default to bash and plan if no tools specified
	if len(allowed) == 0 {
		allowed = []string{"bash", "plan"}
	}

	var tools []fantasy.AgentTool
	for _, name := range allowed {
		switch name {
		case "bash":
			tools = append(tools, NewBashTool(baseDir))
		case "plan":
			tools = append(tools, NewPlanTool())
		case "memory":
			tools = append(tools, NewMemoryToolWithQueue(memQueue))
		// agent_call is added separately when needed
		}
	}

	return FantasyToolSet{tools: tools, allowedList: allowed, baseDir: baseDir, memoryQueue: memQueue}
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
	absTarget, err := filepath.Abs(filepath.Join(absBase, workingDirArg))
	if err != nil {
		return "", err
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

// MemoryParams defines the parameters for the memory tool.
type MemoryParams struct {
	Operation string `json:"operation" description:"The memory operation to perform: 'search', 'store', 'list', 'forget'"`
	Query     string `json:"query,omitempty" description:"For 'search': the semantic search query"`
	Content   string `json:"content,omitempty" description:"For 'store': the memory content to store"`
	Category  string `json:"category,omitempty" description:"For 'store': the memory category (preference, fact, correction, pattern). Default: fact"`
	ID        string `json:"id,omitempty" description:"For 'forget': the memory ID (or prefix) to forget"`
	Limit     int    `json:"limit,omitempty" description:"For 'search' or 'list': maximum number of results. Default: 10"`
}

// NewMemoryTool creates the memory tool for Fantasy (sync mode, shells out).
// Deprecated: Use NewMemoryToolWithQueue for async store operations.
func NewMemoryTool(ayoBinary string) fantasy.AgentTool {
	return NewMemoryToolWithQueue(nil)
}

// NewMemoryToolWithQueue creates the memory tool with optional async queue for store operations.
// If queue is nil, all operations shell out synchronously.
// If queue is provided, store operations are async; search/list/forget remain sync.
func NewMemoryToolWithQueue(queue *memory.Queue) fantasy.AgentTool {
	// Get the binary path for sync operations
	ayoBinary := ""
	if exe, err := os.Executable(); err == nil {
		ayoBinary = exe
	} else {
		ayoBinary = "ayo"
	}

	return fantasy.NewAgentTool(
		"memory",
		"Manage persistent memories that persist across sessions. Use 'search' to find relevant memories, 'store' to save new information, 'list' to see all memories, or 'forget' to remove memories.",
		func(ctx context.Context, params MemoryParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			// Handle store operation with async queue if available
			if params.Operation == "store" {
				if params.Content == "" {
					return fantasy.NewTextErrorResponse("content is required for store operation"), nil
				}

				if queue != nil {
					// Async path: enqueue and return immediately
					category := memory.CategoryFact // default
					switch params.Category {
					case "preference":
						category = memory.CategoryPreference
					case "correction":
						category = memory.CategoryCorrection
					case "pattern":
						category = memory.CategoryPattern
					}

					reqID := queue.Enqueue(params.Content, category, "", "")
					return fantasy.NewTextResponse(fmt.Sprintf(`{"queued":true,"request_id":"%s","message":"Memory queued for storage"}`, reqID)), nil
				}
				// Fall through to sync path if no queue
			}

			// Sync path: shell out to ayo CLI
			var args []string

			switch params.Operation {
			case "search":
				if params.Query == "" {
					return fantasy.NewTextErrorResponse("query is required for search operation"), nil
				}
				args = []string{"memory", "search", params.Query, "--json"}
				if params.Limit > 0 {
					args = append(args, "-n", fmt.Sprintf("%d", params.Limit))
				}

			case "store":
				// Only reached if queue is nil
				args = []string{"memory", "store", params.Content, "--json"}
				if params.Category != "" {
					args = append(args, "-c", params.Category)
				}

			case "list":
				args = []string{"memory", "list", "--json"}
				if params.Limit > 0 {
					args = append(args, "-n", fmt.Sprintf("%d", params.Limit))
				}

			case "forget":
				if params.ID == "" {
					return fantasy.NewTextErrorResponse("id is required for forget operation"), nil
				}
				args = []string{"memory", "forget", params.ID, "--json", "-f"}

			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("unknown operation: %s. Valid operations: search, store, list, forget", params.Operation)), nil
			}

			// Execute the ayo command
			timeout := 30 * time.Second
			execCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			cmd := exec.CommandContext(execCtx, ayoBinary, args...)

			stdoutBuf := &fantasyLimitedBuffer{max: fantasyOutputLimitBytes}
			stderrBuf := &fantasyLimitedBuffer{max: fantasyOutputLimitBytes}
			cmd.Stdout = stdoutBuf
			cmd.Stderr = stderrBuf

			runErr := cmd.Run()

			if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
				return fantasy.NewTextErrorResponse("memory operation timed out"), nil
			}

			if runErr != nil {
				errMsg := stderrBuf.String()
				if errMsg == "" {
					errMsg = runErr.Error()
				}
				return fantasy.NewTextErrorResponse(fmt.Sprintf("memory operation failed: %s", errMsg)), nil
			}

			return fantasy.NewTextResponse(stdoutBuf.String()), nil
		},
	)
}
