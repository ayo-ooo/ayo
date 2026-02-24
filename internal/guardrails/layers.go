// Package guardrails provides safety guardrails for AI agents.
// Guardrails are organized into layers, each providing different types of safety.
package guardrails

// Layer represents a guardrail layer.
type Layer int

const (
	// LayerInfrastructure provides sandbox-based isolation.
	// - Filesystem isolation (read-only host mount)
	// - Network controls
	// - Process isolation
	// - Resource limits
	// This layer is always active when a sandbox is available.
	LayerInfrastructure Layer = iota + 1

	// LayerProtocol provides daemon-enforced safety.
	// - file_request approval flow
	// - Audit logging
	// - Trust level enforcement
	// - Adversarial input detection
	// This layer is always active.
	LayerProtocol

	// LayerPrompt provides prompt-based guidance.
	// - Sandwich pattern (prefix/suffix)
	// - Per-agent guardrails
	// - Squad constitution injection
	// This layer can be configured or disabled.
	LayerPrompt

	// LayerBehavioral provides runtime behavioral controls.
	// - Output filters
	// - Rate limiting
	// - Human-in-the-loop
	// This layer is optional.
	LayerBehavioral
)

// String returns the layer name.
func (l Layer) String() string {
	switch l {
	case LayerInfrastructure:
		return "infrastructure"
	case LayerProtocol:
		return "protocol"
	case LayerPrompt:
		return "prompt"
	case LayerBehavioral:
		return "behavioral"
	default:
		return "unknown"
	}
}

// Level represents a guardrail configuration level.
type Level string

const (
	// LevelMinimal enables only L1 (infrastructure) + L2 (protocol).
	// Use for experienced users who want maximum flexibility.
	LevelMinimal Level = "minimal"

	// LevelStandard enables L1 + L2 + L3 (prompt).
	// This is the default for most users.
	LevelStandard Level = "standard"

	// LevelStrict enables all layers including L4 (behavioral).
	// Use for high-risk environments or untrusted agents.
	LevelStrict Level = "strict"
)

// ActiveLayers returns which layers are active for a given level.
func ActiveLayers(level Level) []Layer {
	switch level {
	case LevelMinimal:
		return []Layer{LayerInfrastructure, LayerProtocol}
	case LevelStandard, "":
		return []Layer{LayerInfrastructure, LayerProtocol, LayerPrompt}
	case LevelStrict:
		return []Layer{LayerInfrastructure, LayerProtocol, LayerPrompt, LayerBehavioral}
	default:
		// Default to standard
		return []Layer{LayerInfrastructure, LayerProtocol, LayerPrompt}
	}
}

// IsLayerActive returns true if a layer is active for the given level.
func IsLayerActive(level Level, layer Layer) bool {
	for _, l := range ActiveLayers(level) {
		if l == layer {
			return true
		}
	}
	return false
}

// Config represents guardrails configuration for an agent.
type Config struct {
	// Enabled controls whether guardrails are applied.
	// Defaults to true. Setting to false disables L3/L4 but L1/L2 remain active.
	Enabled *bool `json:"enabled,omitempty"`

	// Level controls which layers are active: "minimal", "standard", or "strict".
	// Defaults to "standard".
	Level Level `json:"level,omitempty"`

	// Sandbox configures infrastructure layer (L1).
	Sandbox *SandboxGuardrails `json:"sandbox,omitempty"`

	// Prompt configures prompt layer (L3).
	Prompt *PromptGuardrails `json:"prompt,omitempty"`

	// Permissions configures protocol layer (L2).
	Permissions *PermissionsGuardrails `json:"permissions,omitempty"`
}

// SandboxGuardrails configures L1 infrastructure controls.
type SandboxGuardrails struct {
	// Network enables network access in sandbox.
	Network *bool `json:"network,omitempty"`

	// Filesystem controls filesystem access: "readonly" or "readwrite".
	Filesystem string `json:"filesystem,omitempty"`
}

// PromptGuardrails configures L3 prompt-based controls.
type PromptGuardrails struct {
	// UseSandwich enables sandwich pattern (prefix/suffix).
	UseSandwich *bool `json:"use_sandwich,omitempty"`

	// CustomPrefix is a path to a custom prefix file.
	CustomPrefix string `json:"custom_prefix,omitempty"`

	// CustomSuffix is a path to a custom suffix file.
	CustomSuffix string `json:"custom_suffix,omitempty"`
}

// PermissionsGuardrails configures L2 protocol controls.
type PermissionsGuardrails struct {
	// AutoApprove enables automatic approval of file modifications.
	AutoApprove *bool `json:"auto_approve,omitempty"`

	// AllowedPaths are glob patterns for paths the agent can access.
	AllowedPaths []string `json:"allowed_paths,omitempty"`

	// DeniedPaths are glob patterns for paths the agent cannot access.
	DeniedPaths []string `json:"denied_paths,omitempty"`
}

// IsEnabled returns true if guardrails are enabled.
func (c *Config) IsEnabled() bool {
	if c == nil || c.Enabled == nil {
		return true // Default to enabled
	}
	return *c.Enabled
}

// GetLevel returns the guardrail level, defaulting to standard.
func (c *Config) GetLevel() Level {
	if c == nil || c.Level == "" {
		return LevelStandard
	}
	return c.Level
}
