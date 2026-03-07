package types

// Config represents the complete configuration for an agent or team executable
type Config struct {
	Agent    AgentConfig    `toml:"agent"`
	CLI      CLIConfig      `toml:"cli"`
	Input    InputConfig    `toml:"input"`
	Output   OutputConfig   `toml:"output"`
	Prompts  PromptsConfig  `toml:"prompts"`
	Triggers TriggersConfig `toml:"triggers,omitempty"`
	Evals    EvalsConfig    `toml:"evals,omitempty"`
}

// AgentConfig contains agent-specific settings
type AgentConfig struct {
	Name        string          `toml:"name"`
	Description string          `toml:"description"`
	Model       string          `toml:"model"`
	Tools       AgentTools      `toml:"tools,omitempty"`
	Memory      AgentMemory     `toml:"memory,omitempty"`
	Sandbox     AgentSandbox    `toml:"sandbox,omitempty"`
}

// AgentTools defines which tools the agent can use
type AgentTools struct {
	Allowed []string `toml:"allowed,omitempty"`
}

// AgentMemory configures agent memory settings
type AgentMemory struct {
	Enabled bool   `toml:"enabled"`
	Scope   string `toml:"scope"` // "agent" | "session" | "global"
}

// AgentSandbox configures sandbox settings
type AgentSandbox struct {
	Network  bool   `toml:"network"`
	HostPath string `toml:"host_path"`
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

// TriggersConfig defines automatic triggers for agent execution
type TriggersConfig struct {
	Watch    []string `toml:"watch,omitempty"`    // Files/directories to watch
	Schedule string   `toml:"schedule,omitempty"` // Cron expression
	Events   []string `toml:"events,omitempty"`  // Event types to respond to
}

// EvalsConfig defines evaluation settings for testing agent outputs
type EvalsConfig struct {
	Enabled         bool   `toml:"enabled"`                    // Whether to run evaluations
	File            string `toml:"file"`                       // Path to evals.csv file
	JudgeModel      string `toml:"judge_model"`               // Model to use for judging
	JudgeProvider   string `toml:"judge_provider"`            // Provider to use for judging
	Criteria        string `toml:"criteria"`                  // Evaluation criteria (e.g., "accuracy,helpfulness")
}
