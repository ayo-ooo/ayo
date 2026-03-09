// Package tools provides tool management for the build system.
package tools

// ToolType represents the type of tool.
type ToolType string

const (
	// Builtin tools that are provided by ayo
	ToolTypeBuiltin ToolType = "builtin"
	
	// External tools that are executable programs
	ToolTypeExternal ToolType = "external"
)

// Tool represents a tool that can be used by agents.
type Tool struct {
	Name        string
	Type        ToolType
	Description string
	Path        string // For external tools, the path to the executable
}

// BuiltinTools lists the tools that are built into ayo.
var BuiltinTools = map[string]Tool{
	"bash": {
		Name:        "bash",
		Type:        ToolTypeBuiltin,
		Description: "Execute shell commands",
	},
	"file_read": {
		Name:        "file_read",
		Type:        ToolTypeBuiltin,
		Description: "Read file contents",
	},
	"file_write": {
		Name:        "file_write",
		Type:        ToolTypeBuiltin,
		Description: "Write to files",
	},
	"git": {
		Name:        "git",
		Type:        ToolTypeBuiltin,
		Description: "Git operations",
	},
}

// GetTool returns a tool by name.
func GetTool(name string) (Tool, bool) {
	if tool, exists := BuiltinTools[name]; exists {
		return tool, true
	}
	return Tool{}, false
}

// IsBuiltinTool returns true if the tool is a builtin tool.
func IsBuiltinTool(name string) bool {
	_, exists := BuiltinTools[name]
	return exists
}

// IsExternalTool returns true if the tool is an external executable.
func IsExternalTool(name string) bool {
	return !IsBuiltinTool(name)
}
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
