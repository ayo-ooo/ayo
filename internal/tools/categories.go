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
