// Package tools provides tool category management and stateful tool abstractions.
package tools

import (
	"github.com/alexcabrera/ayo/internal/config"
)

// ExecutionContext specifies where a tool should execute.
type ExecutionContext string

const (
	// ExecHost means the tool executes on the host machine.
	// Examples: memory search, agent_call, filesystem browsing.
	ExecHost ExecutionContext = "host"

	// ExecSandbox means the tool executes inside the sandbox.
	// Examples: bash, code execution.
	ExecSandbox ExecutionContext = "sandbox"

	// ExecBridge means the tool needs access to both host and sandbox.
	// Examples: file_request (host->sandbox), publish (sandbox->host).
	ExecBridge ExecutionContext = "bridge"
)

// ToolExecContext maps tool names to their execution context.
// Tools not listed default to ExecSandbox when a sandbox is available.
var ToolExecContext = map[string]ExecutionContext{
	// Host-side tools - these access host services or filesystem
	"memory":     ExecHost,
	"agent_call": ExecHost,
	"delegate":   ExecHost,
	"todo":       ExecHost,

	// Sandbox tools - these execute commands inside the sandbox
	"bash": ExecSandbox,

	// Bridge tools - these need access to both host and sandbox
	"file_request": ExecBridge,
	"publish":      ExecBridge,
}

// GetExecutionContext returns the execution context for a tool.
// If the tool is not explicitly mapped, returns ExecSandbox as default.
func GetExecutionContext(toolName string) ExecutionContext {
	if ctx, ok := ToolExecContext[toolName]; ok {
		return ctx
	}
	return ExecSandbox
}

// IsHostTool returns true if the tool should execute on the host.
func IsHostTool(toolName string) bool {
	return GetExecutionContext(toolName) == ExecHost
}

// IsSandboxTool returns true if the tool should execute in the sandbox.
func IsSandboxTool(toolName string) bool {
	return GetExecutionContext(toolName) == ExecSandbox
}

// IsBridgeTool returns true if the tool needs both host and sandbox access.
func IsBridgeTool(toolName string) bool {
	return GetExecutionContext(toolName) == ExecBridge
}

// Category represents a semantic tool slot that can be filled by different implementations.
// Categories allow users to swap implementations without changing agent configurations.
type Category string

// Defined tool categories.
const (
	// CategoryPlan is for durable project planning that persists across sessions.
	// No default - must be provided by plugin (e.g., ticket).
	// Note: The built-in "todo" tool is always available separately (not via category).
	CategoryPlan Category = "plan"

	// CategoryShell is for command execution.
	// Default: "bash"
	CategoryShell Category = "shell"

	// CategorySearch is for web search capabilities.
	// No default - must be provided by plugin.
	CategorySearch Category = "search"
)

// builtinDefaults maps categories to their default tool implementations.
// These are the tools that ship with ayo.
// Note: CategoryPlan and CategorySearch have no defaults - they require plugins.
var builtinDefaults = map[Category]string{
	CategoryShell: "bash",
}

// knownCategories lists all defined categories (with or without defaults).
var knownCategories = map[Category]bool{
	CategoryPlan:   true,
	CategoryShell:  true,
	CategorySearch: true,
}

// IsCategory returns true if the name is a known tool category.
func IsCategory(name string) bool {
	return knownCategories[Category(name)]
}

// DefaultForCategory returns the default tool implementation for a category.
// Returns empty string if the category has no default.
func DefaultForCategory(cat Category) string {
	return builtinDefaults[cat]
}

// ResolveToolName converts a category or tool name to a concrete tool name.
// Resolution order:
//  1. If name is a category, check config.DefaultTools for user override
//  2. If name is a category with builtin default, use that
//  3. If name is a category with no default, return empty string (tool not loaded)
//  4. If name is not a category, check config.DefaultTools for alias
//  5. Otherwise return name as-is
func ResolveToolName(name string, cfg *config.Config) string {
	cat := Category(name)

	// Check if it's a known category
	if knownCategories[cat] {
		// Check for user override in config
		if cfg != nil && cfg.DefaultTools != nil {
			if override, ok := cfg.DefaultTools[name]; ok && override != "" {
				return override
			}
		}
		// Return builtin default (may be empty for categories like plan/search)
		return builtinDefaults[cat]
	}

	// Not a category - check for alias in config
	if cfg != nil && cfg.DefaultTools != nil {
		if resolved, ok := cfg.DefaultTools[name]; ok && resolved != "" {
			return resolved
		}
	}

	return name
}

// ListCategories returns all defined categories and their defaults.
// Categories without defaults will have empty string values.
func ListCategories() map[Category]string {
	result := make(map[Category]string, len(knownCategories))
	for cat := range knownCategories {
		result[cat] = builtinDefaults[cat] // Will be empty for cats without defaults
	}
	return result
}
