package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/capabilities"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/plugins"
	"github.com/alexcabrera/ayo/internal/sandbox"
	"github.com/alexcabrera/ayo/internal/share"
	"github.com/alexcabrera/ayo/internal/squads"
	"github.com/alexcabrera/ayo/internal/tools"
	"github.com/alexcabrera/ayo/internal/tools/delegate"
	"github.com/alexcabrera/ayo/internal/tools/findagent"
	"github.com/alexcabrera/ayo/internal/tools/requestaccess"
)

// Tool parameter types for Fantasy

// BashParams is an alias to sandbox.BashParams.
type BashParams = sandbox.BashParams

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

			result := CommandResult{
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



// FantasyToolSet wraps Fantasy tools for use with the runner.
type FantasyToolSet struct {
	tools           []fantasy.AgentTool
	allowedList     []string
	baseDir         string
	memoryQueue     *memory.Queue           // Optional async memory queue
	depth           int
	sandboxExecutor *sandbox.Executor       // Optional sandbox executor for bash
	statefulTools   []tools.StatefulTool    // Tools that need cleanup
}

// ToolSetOptions configures the Fantasy tool set.
type ToolSetOptions struct {
	AllowedTools       []string
	BaseDir            string
	MemoryQueue        *memory.Queue
	Depth              int
	SandboxExecutor    *sandbox.Executor
	CapabilitySearcher *capabilities.CapabilitySearcher // Optional for find_agent tool
	PlannerTools       []fantasy.AgentTool              // Tools from planner plugins
	ShareService       *share.Service                   // Optional for request_access tool
	SessionID          string                           // Session ID for session-scoped shares

	// Squad delegation support
	SquadName         string                // Squad name for delegation context
	SquadAgents       []string              // Agent handles in the squad
	SquadConstitution *squads.Constitution  // Squad constitution for injection
	SquadInvoker      squads.AgentInvoker   // Invoker for agent delegation
}

// NewFantasyToolSetWithOptions creates a Fantasy tool set with all options.
// Deprecated: Use NewFantasyToolSet with ToolSetOptions directly.
// Task management tools are now provided by planners via PlannerTools option.
func NewFantasyToolSetWithOptions(allowed []string, baseDir string, memQueue *memory.Queue, depth int, _ bool) FantasyToolSet {
	return NewFantasyToolSet(ToolSetOptions{
		AllowedTools: allowed,
		BaseDir:      baseDir,
		MemoryQueue:  memQueue,
		Depth:        depth,
	})
}

// NewFantasyToolSet creates a Fantasy tool set with configurable options.
// This is the recommended way to create a tool set with sandbox support.
func NewFantasyToolSet(opts ToolSetOptions) FantasyToolSet {
	baseDir := opts.BaseDir
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}

	// Load config for category resolution
	cfg, _ := config.Load(config.DefaultPath())

	var fantasyTools []fantasy.AgentTool
	loadedTools := make(map[string]bool)

	// Track stateful tools that need cleanup
	var statefulTools []tools.StatefulTool

	// Default to bash if no tools specified
	allowed := opts.AllowedTools
	if len(allowed) == 0 {
		allowed = []string{"bash"}
	}

	for _, name := range allowed {
		// Skip old planning names - task management now via planners
		if name == "todo" || name == "todos" || name == "planning" {
			continue
		}

		// Resolve category to concrete tool name
		resolvedName := tools.ResolveToolName(name, &cfg)

		// Skip if empty (category with no configured default)
		if resolvedName == "" {
			continue
		}

		// Skip if already loaded (handles aliases pointing to same tool)
		if loadedTools[resolvedName] {
			continue
		}

		switch resolvedName {
		case "bash":
			// Use sandbox executor if provided, otherwise use local execution
			if opts.SandboxExecutor != nil {
				fantasyTools = append(fantasyTools, sandbox.NewBashTool(opts.SandboxExecutor))
			} else {
				fantasyTools = append(fantasyTools, NewBashTool(baseDir))
			}
			loadedTools[resolvedName] = true
		case "memory":
			fantasyTools = append(fantasyTools, NewMemoryToolWithQueue(opts.MemoryQueue))
			loadedTools[resolvedName] = true
		case "find_agent":
			// Only add if capability searcher is configured
			if opts.CapabilitySearcher != nil {
				fantasyTools = append(fantasyTools, findagent.NewFindAgentTool(findagent.ToolConfig{
					Searcher: opts.CapabilitySearcher,
				}))
				loadedTools[resolvedName] = true
			}
		case "request_access":
			// Only add if share service is configured
			if opts.ShareService != nil {
				fantasyTools = append(fantasyTools, requestaccess.NewRequestAccessTool(requestaccess.ToolConfig{
					ShareService:  opts.ShareService,
					SessionID:     opts.SessionID,
					SessionScoped: nil, // Default to true (session-scoped shares)
				}))
				loadedTools[resolvedName] = true
			}
		case "delegate":
			// Only add if squad invoker is configured
			if opts.SquadInvoker != nil && opts.SquadName != "" {
				fantasyTools = append(fantasyTools, delegate.NewDelegateTool(delegate.ToolConfig{
					SquadName:    opts.SquadName,
					SquadAgents:  opts.SquadAgents,
					Constitution: opts.SquadConstitution,
					Invoker:      opts.SquadInvoker,
				}))
				loadedTools[resolvedName] = true
			}
		default:
			// Try to load as external tool from plugins
			if tool := loadExternalTool(resolvedName, baseDir, opts.Depth, &cfg); tool != nil {
				fantasyTools = append(fantasyTools, tool)
				loadedTools[resolvedName] = true
			}
		}
	}

	// Add planner tools (from SandboxPlanners.NearTerm.Tools() and LongTerm.Tools())
	// These are always-available like todo, not subject to allowedTools filtering
	for _, plannerTool := range opts.PlannerTools {
		toolName := plannerTool.Info().Name
		// Skip if a tool with this name is already loaded (avoid collisions)
		if !loadedTools[toolName] {
			fantasyTools = append(fantasyTools, plannerTool)
			loadedTools[toolName] = true
		}
	}

	// Add delegate tool automatically when squad context is configured
	// This allows squad agents to delegate work to other agents without
	// explicitly listing "delegate" in AllowedTools
	if opts.SquadInvoker != nil && opts.SquadName != "" && !loadedTools["delegate"] {
		fantasyTools = append(fantasyTools, delegate.NewDelegateTool(delegate.ToolConfig{
			SquadName:    opts.SquadName,
			SquadAgents:  opts.SquadAgents,
			Constitution: opts.SquadConstitution,
			Invoker:      opts.SquadInvoker,
		}))
		loadedTools["delegate"] = true
	}

	return FantasyToolSet{
		tools:           fantasyTools,
		allowedList:     allowed,
		baseDir:         baseDir,
		memoryQueue:     opts.MemoryQueue,
		depth:           opts.Depth,
		sandboxExecutor: opts.SandboxExecutor,
		statefulTools:   statefulTools,
	}
}

// Tools returns the Fantasy agent tools.
func (ts FantasyToolSet) Tools() []fantasy.AgentTool {
	return ts.tools
}

// HasTool checks if a tool name is in the allowed list.
func (ts FantasyToolSet) HasTool(name string) bool {
	return slices.Contains(ts.allowedList, name)
}

// Close releases resources held by stateful tools.
func (ts *FantasyToolSet) Close() error {
	var lastErr error
	for _, tool := range ts.statefulTools {
		if err := tool.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// GetPlannerTools extracts tools from SandboxPlanners for use with ToolSetOptions.
// This collects tools from both near-term and long-term planners.
func GetPlannerTools(nearTerm, longTerm interface{ Tools() []fantasy.AgentTool }) []fantasy.AgentTool {
	var tools []fantasy.AgentTool

	if nearTerm != nil {
		tools = append(tools, nearTerm.Tools()...)
	}
	if longTerm != nil {
		tools = append(tools, longTerm.Tools()...)
	}

	return tools
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

// MemoryParams defines the parameters for the memory tool.
type MemoryParams struct {
	Operation string `json:"operation" description:"The memory operation to perform: 'search', 'store', 'list', 'forget', 'topics', 'link'"`
	Query     string `json:"query,omitempty" description:"For 'search': the semantic search query"`
	Content   string `json:"content,omitempty" description:"For 'store': the memory content to store"`
	Category  string `json:"category,omitempty" description:"For 'store': the memory category (preference, fact, correction, pattern). Default: fact"`
	ID        string `json:"id,omitempty" description:"For 'forget' or 'link': the first memory ID"`
	TargetID  string `json:"target_id,omitempty" description:"For 'link': the second memory ID to link with"`
	Limit     int    `json:"limit,omitempty" description:"For 'search' or 'list': maximum number of results. Default: 10"`
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
		"Manage persistent memories that persist across sessions. Use 'search' to find relevant memories, 'store' to save new information, 'list' to see all memories, 'forget' to remove memories, 'topics' to list all memory topics, or 'link' to connect two related memories.",
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

			case "topics":
				args = []string{"memory", "topics", "--json"}

			case "link":
				if params.ID == "" {
					return fantasy.NewTextErrorResponse("id is required for link operation"), nil
				}
				if params.TargetID == "" {
					return fantasy.NewTextErrorResponse("target_id is required for link operation"), nil
				}
				args = []string{"memory", "link", params.ID, params.TargetID, "--json"}

			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("unknown operation: %s. Valid operations: search, store, list, forget, topics, link", params.Operation)), nil
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

// loadExternalTool attempts to load a tool from installed plugins.
// Returns nil if the tool is not found.
// It first checks if the toolName is a tool alias (e.g., "search") and resolves it
// to the configured default tool (e.g., "searxng").
func loadExternalTool(toolName string, baseDir string, depth int, cfg *config.Config) fantasy.AgentTool {
	// First, check if this is a tool alias that should be resolved
	resolvedName := resolveToolAlias(toolName, cfg)

	// Load plugin registry
	registry, err := plugins.LoadRegistry()
	if err != nil {
		return nil
	}

	// Search all enabled plugins for this tool
	for _, plugin := range registry.ListEnabled() {
		for _, tool := range plugin.Tools {
			if tool == resolvedName {
				// Load tool definition
				def, err := plugins.LoadToolDefinition(plugin.Path, resolvedName)
				if err != nil {
					continue
				}

				return NewExternalTool(def, plugin.Path, baseDir, depth)
			}
		}
	}

	return nil
}

// resolveToolAlias checks if the given tool name is an alias and returns the
// configured concrete tool name. Returns the original name if no alias is configured.
func resolveToolAlias(toolName string, cfg *config.Config) string {
	if cfg == nil || cfg.DefaultTools == nil {
		return toolName
	}

	if resolved, ok := cfg.DefaultTools[toolName]; ok && resolved != "" {
		return resolved
	}

	return toolName
}
