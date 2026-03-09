package types

// Config represents the configuration for an agent executable
type Config struct {
	Agent    AgentConfig    `toml:"agent"`
	CLI      CLIConfig      `toml:"cli"`
	Input    InputConfig    `toml:"input,omitempty"`
	Output   OutputConfig   `toml:"output,omitempty"`
	Prompts  PromptsConfig  `toml:"prompts,omitempty"`
}

// AgentConfig contains agent-specific settings
type AgentConfig struct {
	Name        string          `toml:"name"`
	Description string          `toml:"description"`
	Model       string          `toml:"model"`
	Tools       AgentTools      `toml:"tools,omitempty"`
	Memory      AgentMemory     `toml:"memory,omitempty"`
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
	Mode        string           `toml:"mode"` // "structured" | "freeform" | "hybrid"
	Description string           `toml:"description"`
	Flags       map[string]CLIFlag `toml:"flags,omitempty"`
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
