// Package tools provides tool category management and stateful tool abstractions.
package tools

import (
	"github.com/alexcabrera/ayo/internal/config"
)

// Category represents a semantic tool slot that can be filled by different implementations.
// Categories allow users to swap implementations without changing agent configurations.
type Category string

// Defined tool categories.
const (
	// CategoryPlanning is for task tracking during execution.
	// Default: "todo" (flat list)
	CategoryPlanning Category = "planning"

	// CategoryShell is for command execution.
	// Default: "bash"
	CategoryShell Category = "shell"

	// CategorySearch is for web search capabilities.
	// No default - must be provided by plugin.
	CategorySearch Category = "search"
)

// builtinDefaults maps categories to their default tool implementations.
// These are the tools that ship with ayo.
var builtinDefaults = map[Category]string{
	CategoryPlanning: "todo",
	CategoryShell:    "bash",
}

// IsCategory returns true if the name is a known tool category.
func IsCategory(name string) bool {
	_, ok := builtinDefaults[Category(name)]
	return ok
}

// DefaultForCategory returns the default tool implementation for a category.
// Returns empty string if the category has no default.
func DefaultForCategory(cat Category) string {
	return builtinDefaults[cat]
}

// ResolveToolName converts a category or tool name to a concrete tool name.
// Resolution order:
//  1. If name is a category, check config.DefaultTools for user override
//  2. If name is a category with no override, use builtin default
//  3. If name is not a category, check config.DefaultTools for alias
//  4. Otherwise return name as-is
func ResolveToolName(name string, cfg *config.Config) string {
	cat := Category(name)

	// Check if it's a known category
	if defaultTool, isCategory := builtinDefaults[cat]; isCategory {
		// Check for user override in config
		if cfg != nil && cfg.DefaultTools != nil {
			if override, ok := cfg.DefaultTools[name]; ok && override != "" {
				return override
			}
		}
		return defaultTool
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
func ListCategories() map[Category]string {
	// Return a copy to prevent modification
	result := make(map[Category]string, len(builtinDefaults))
	for k, v := range builtinDefaults {
		result[k] = v
	}
	return result
}
