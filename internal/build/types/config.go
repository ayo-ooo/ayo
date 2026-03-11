package types

import (
	"fmt"
	"strings"
)

// Config represents the configuration for an agent executable
type Config struct {
	Agent    AgentConfig    `toml:"agent"`
	CLI      CLIConfig      `toml:"cli"`
	Input    InputConfig    `toml:"input,omitempty"`
	Output   OutputConfig   `toml:"output,omitempty"`
	Prompts  PromptsConfig  `toml:"prompts,omitempty"`
	Build    BuildConfig    `toml:"build,omitempty"`
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	if err := c.Agent.Validate(); err != nil {
		return fmt.Errorf("agent config: %w", err)
	}
	if err := c.CLI.Validate(); err != nil {
		return fmt.Errorf("cli config: %w", err)
	}
	return nil
}

// AgentConfig contains agent-specific settings
type AgentConfig struct {
	Name        string      `toml:"name"`
	Description string      `toml:"description"`
	Model       string      `toml:"model"`
	Tools       AgentTools  `toml:"tools,omitempty"`
	Memory      AgentMemory `toml:"memory,omitempty"`
	Temperature *float64    `toml:"temperature,omitempty"` // Model temperature (0.0 - 2.0)
	MaxTokens   int         `toml:"max_tokens,omitempty"`   // Maximum tokens for response
}

// Validate validates the agent configuration.
func (a *AgentConfig) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("name is required")
	}
	if a.Description == "" {
		return fmt.Errorf("description is required")
	}
	if a.Model == "" {
		return fmt.Errorf("model is required")
	}
	if a.Temperature != nil && (*a.Temperature < 0.0 || *a.Temperature > 2.0) {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}
	if a.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be non-negative")
	}
	if a.Memory.Enabled {
		if a.Memory.Scope != "agent" && a.Memory.Scope != "session" {
			return fmt.Errorf("memory scope must be 'agent' or 'session', got '%s'", a.Memory.Scope)
		}
	}
	return nil
}

// AgentTools defines which tools the agent can use
type AgentTools struct {
	Allowed []string `toml:"allowed,omitempty"`
}

// AgentMemory configures agent memory settings
type AgentMemory struct {
	Enabled bool   `toml:"enabled"`
	Scope   string `toml:"scope"` // "agent" | "session"
}

// CLIConfig defines the command-line interface for the executable
type CLIConfig struct {
	Mode        string            `toml:"mode"` // "structured" | "freeform" | "hybrid"
	Description string            `toml:"description"`
	Flags       map[string]CLIFlag `toml:"flags,omitempty"`
}

// Validate validates the CLI configuration.
func (c *CLIConfig) Validate() error {
	if c.Mode == "" {
		return fmt.Errorf("mode is required")
	}
	validModes := map[string]bool{
		"structured": true,
		"freeform":   true,
		"hybrid":     true,
	}
	if !validModes[c.Mode] {
		return fmt.Errorf("invalid mode '%s': must be 'structured', 'freeform', or 'hybrid'", c.Mode)
	}
	if c.Description == "" {
		return fmt.Errorf("description is required")
	}
	for name, flag := range c.Flags {
		if err := flag.Validate(); err != nil {
			return fmt.Errorf("flag '%s': %w", name, err)
		}
	}
	return nil
}

// CLIFlag defines a single command-line flag
type CLIFlag struct {
	Name        string `toml:"name"`
	Type        string `toml:"type"` // "string" | "int" | "float" | "bool" | "array"
	Short       string `toml:"short,omitempty"`
	Position    int    `toml:"position,omitempty"` // Positional argument index
	Required    bool   `toml:"required"`
	Multiple    bool   `toml:"multiple"` // Accept multiple values
	Description string `toml:"description"`
	Default     any    `toml:"default,omitempty"`
}

// Validate validates the CLI flag.
func (f *CLIFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name is required")
	}
	if f.Description == "" {
		return fmt.Errorf("description is required")
	}
	validTypes := map[string]bool{
		"string": true,
		"int":    true,
		"float":  true,
		"bool":   true,
		"array":  true,
	}
	if !validTypes[f.Type] {
		return fmt.Errorf("invalid type '%s': must be 'string', 'int', 'float', 'bool', or 'array'", f.Type)
	}
	if len(f.Short) > 1 {
		return fmt.Errorf("short flag must be a single character, got '%s'", f.Short)
	}
	if f.Position < 0 {
		return fmt.Errorf("position must be non-negative")
	}
	return nil
}

// InputConfig defines the expected input schema
type InputConfig struct {
	Schema map[string]any `toml:"schema"`
}

// OutputConfig defines the expected output schema
type OutputConfig struct {
	Schema map[string]any `toml:"schema"`
}

// PromptsConfig contains prompt templates
type PromptsConfig struct {
	System string `toml:"system,omitempty"`
	User   string `toml:"user,omitempty"`
}

// BuildConfig contains build configuration settings
type BuildConfig struct {
	Targets []BuildTarget `toml:"targets,omitempty"`
}

// BuildTarget defines a build target platform
type BuildTarget struct {
	OS   string `toml:"os"`
	Arch string `toml:"arch"`
}

// IsValidModel checks if a model string is supported.
func IsValidModel(model string) bool {
	// Check for common model prefixes
	modelLower := strings.ToLower(model)
	return strings.HasPrefix(modelLower, "gpt-") ||
		strings.HasPrefix(modelLower, "claude-") ||
		strings.HasPrefix(modelLower, "o1-") ||
		strings.HasPrefix(modelLower, "gemini-")
}
