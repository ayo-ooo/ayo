package run

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"charm.land/fantasy"

	"github.com/alexcabrera/ayo/internal/plugins"
	uipkg "github.com/alexcabrera/ayo/internal/ui"
)

// externalToolWrapper wraps an external tool to implement the AgentTool interface.
// This is needed because Fantasy's NewAgentTool uses generics with schema generation,
// but external tools have dynamic schemas defined in JSON.
type externalToolWrapper struct {
	def       *plugins.ToolDefinition
	pluginDir string
	baseDir   string
	depth     int
}

func (w *externalToolWrapper) Info() fantasy.ToolInfo {
	return fantasy.ToolInfo{
		Name:        w.def.Name,
		Description: w.def.Description,
		Parameters:  w.def.ToJSONSchema(),
	}
}

func (w *externalToolWrapper) Run(ctx context.Context, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	// Parse parameters from call
	var params map[string]any
	if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid parameters: %v", err)), nil
	}

	return executeExternalTool(ctx, w.def, w.pluginDir, w.baseDir, params, w.depth)
}

func (w *externalToolWrapper) IsParallel() bool {
	return false // External tools run sequentially by default
}

func (w *externalToolWrapper) ProviderOptions() fantasy.ProviderOptions {
	return fantasy.ProviderOptions{}
}

func (w *externalToolWrapper) SetProviderOptions(opts fantasy.ProviderOptions) {
	// Not used for external tools
}

// NewExternalTool creates a Fantasy tool from a plugin tool definition.
func NewExternalTool(def *plugins.ToolDefinition, pluginDir string, baseDir string, depth int) fantasy.AgentTool {
	return &externalToolWrapper{
		def:       def,
		pluginDir: pluginDir,
		baseDir:   baseDir,
		depth:     depth,
	}
}

// executeExternalTool runs an external tool based on its definition.
func executeExternalTool(
	ctx context.Context,
	def *plugins.ToolDefinition,
	pluginDir string,
	baseDir string,
	params map[string]any,
	depth int,
) (fantasy.ToolResponse, error) {
	// Validate required parameters
	for _, param := range def.Parameters {
		if param.Required {
			val, exists := params[param.Name]
			if !exists || val == nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("parameter '%s' is required", param.Name)), nil
			}
			// Check for empty strings
			if str, ok := val.(string); ok && strings.TrimSpace(str) == "" {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("parameter '%s' cannot be empty", param.Name)), nil
			}
		}
	}

	// Check dependencies
	for _, binary := range def.DependsOn {
		if _, err := exec.LookPath(binary); err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("required binary not found: %s", binary)), nil
		}
	}

	// Resolve the command
	command := def.Command
	commandPath, err := exec.LookPath(command)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("command not found: %s", command)), nil
	}

	// Build arguments
	args, err := buildExternalToolArgs(def, params)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("build arguments: %v", err)), nil
	}

	// Resolve working directory
	workingDir := baseDir
	switch def.WorkingDir {
	case "plugin":
		workingDir = pluginDir
	case "param":
		if wd, ok := params["working_dir"].(string); ok && wd != "" {
			resolved, err := fantasyResolveWorkingDir(baseDir, wd)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid working_dir: %v", err)), nil
			}
			workingDir = resolved
		}
	case "inherit", "":
		// Use baseDir (default)
	}

	// Set up timeout
	timeout := fantasyDefaultToolTimeout
	if def.Timeout > 0 {
		timeout = time.Duration(def.Timeout) * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build command
	cmd := exec.CommandContext(execCtx, commandPath, args...)
	cmd.Dir = workingDir

	// Set environment variables
	if len(def.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range def.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Capture output
	stdoutBuf := &fantasyLimitedBuffer{max: fantasyOutputLimitBytes * 2}
	stderrBuf := &fantasyLimitedBuffer{max: fantasyOutputLimitBytes}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	// Show spinner unless quiet mode
	var spinner *uipkg.CrushSpinner
	if !def.Quiet {
		spinner = uipkg.NewCrushSpinnerWithDepth(def.Name, depth)
		spinner.Start()
	}

	// Run command
	runErr := cmd.Run()

	// Stop spinner
	if spinner != nil {
		if runErr != nil {
			spinner.StopWithError(def.Name + " failed")
		} else {
			spinner.Stop()
		}
	}

	// Build result
	result := externalToolResult{
		Stdout:    stdoutBuf.String(),
		Stderr:    stderrBuf.String(),
		Truncated: stdoutBuf.truncated || stderrBuf.truncated,
	}

	if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
		result.ExitCode = -1
		result.Error = fmt.Sprintf("%s timed out", def.Name)
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
}

// buildExternalToolArgs builds command line arguments from the tool definition and parameters.
func buildExternalToolArgs(def *plugins.ToolDefinition, params map[string]any) ([]string, error) {
	var args []string

	// Start with default args from definition
	for _, arg := range def.Args {
		expanded := expandArgTemplate(arg, params)
		if expanded != "" {
			args = append(args, expanded)
		}
	}

	// Collect positional arguments
	type positionalArg struct {
		position int
		value    string
	}
	var positionalArgs []positionalArg

	// Process each parameter
	for _, param := range def.Parameters {
		val, exists := params[param.Name]
		if !exists || val == nil {
			continue
		}

		// Skip if empty and OmitIfEmpty is set
		if param.OmitIfEmpty && isEmptyValue(val) {
			continue
		}

		// Handle positional arguments
		if param.Position != nil {
			strVal := formatValue(val)
			if strVal != "" {
				positionalArgs = append(positionalArgs, positionalArg{
					position: *param.Position,
					value:    strVal,
				})
			}
			continue
		}

		// Handle flag arguments
		if param.ArgTemplate != "" {
			expanded := expandParamTemplate(param.ArgTemplate, param.Name, val)
			if expanded != "" {
				// Handle templates that produce multiple args (e.g., "--flag value")
				parts := splitArgs(expanded)
				args = append(args, parts...)
			}
		} else {
			// Default behavior based on type
			switch param.Type {
			case "boolean":
				if boolVal, ok := val.(bool); ok && boolVal {
					args = append(args, fmt.Sprintf("--%s", param.Name))
				}
			default:
				strVal := formatValue(val)
				if strVal != "" {
					args = append(args, fmt.Sprintf("--%s=%s", param.Name, strVal))
				}
			}
		}
	}

	// Sort and add positional arguments
	for i := 0; i < len(positionalArgs); i++ {
		for _, pa := range positionalArgs {
			if pa.position == i {
				args = append(args, pa.value)
				break
			}
		}
	}

	return args, nil
}

// expandArgTemplate expands {{param}} placeholders in an argument template.
func expandArgTemplate(template string, params map[string]any) string {
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		name := match[2 : len(match)-2] // Extract param name
		if val, ok := params[name]; ok {
			return formatValue(val)
		}
		return ""
	})
}

// expandParamTemplate expands a parameter's arg template.
func expandParamTemplate(template string, paramName string, value any) string {
	// Replace {{value}} with the actual value
	strVal := formatValue(value)
	result := strings.ReplaceAll(template, "{{value}}", strVal)
	result = strings.ReplaceAll(result, "{{name}}", paramName)
	return result
}

// formatValue converts a value to a string representation.
func formatValue(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		// JSON numbers are float64
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case int, int64, int32:
		return fmt.Sprintf("%d", v)
	case []any:
		// For arrays, join with comma
		var parts []string
		for _, item := range v {
			parts = append(parts, formatValue(item))
		}
		return strings.Join(parts, ",")
	default:
		// Try JSON marshaling for complex types
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(data)
	}
}

// isEmptyValue checks if a value should be considered empty.
func isEmptyValue(val any) bool {
	switch v := val.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(v) == ""
	case bool:
		return !v
	case float64:
		return v == 0
	case int, int64, int32:
		return v == 0
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

// splitArgs splits a string into arguments, respecting quotes.
func splitArgs(s string) []string {
	var args []string
	var current bytes.Buffer
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuote {
			if c == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(c)
			}
		} else {
			switch c {
			case '"', '\'':
				inQuote = true
				quoteChar = c
			case ' ', '\t':
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			default:
				current.WriteByte(c)
			}
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// externalToolResult holds the result of an external tool invocation.
type externalToolResult struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr,omitempty"`
	ExitCode  int    `json:"exit_code"`
	TimedOut  bool   `json:"timed_out,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (r externalToolResult) String() string {
	// For successful runs, just return stdout
	if r.ExitCode == 0 && r.Error == "" && !r.TimedOut {
		if r.Stdout != "" {
			return r.Stdout
		}
		return "[command completed successfully with no output]"
	}

	// For errors, return structured JSON
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf(`{"error":"marshal error: %v"}`, err)
	}
	return string(data)
}
