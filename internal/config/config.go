package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexcabrera/ayo/internal/paths"

	"charm.land/catwalk/pkg/catwalk"
)

// ModelType represents the type of model (large for main inference, small for fast tasks).
type ModelType string

const (
	// ModelTypeLarge is the primary model for main inference tasks.
	ModelTypeLarge ModelType = "large"
	// ModelTypeSmall is a fast/cheap model for internal tasks (titles, memory extraction).
	ModelTypeSmall ModelType = "small"
)

// SelectedModel represents a fully configured model selection.
// This allows specifying provider, model ID, and model-specific parameters.
type SelectedModel struct {
	// Model is the model ID (e.g., "claude-sonnet-4-5-20250929", "gpt-4o").
	Model string `json:"model"`

	// Provider is the provider ID (e.g., "anthropic", "openai").
	// If empty, the model ID is used to find a matching provider.
	Provider string `json:"provider,omitempty"`

	// ReasoningEffort controls reasoning depth for models that support it (OpenAI o1/o3).
	// Valid values: "low", "medium", "high".
	ReasoningEffort string `json:"reasoning_effort,omitempty"`

	// Think enables extended thinking mode for Anthropic Claude models.
	Think bool `json:"think,omitempty"`

	// MaxTokens overrides the default max tokens for this model.
	MaxTokens int64 `json:"max_tokens,omitempty"`

	// Temperature controls randomness (0.0-2.0). If nil, uses model default.
	Temperature *float64 `json:"temperature,omitempty"`

	// TopP controls nucleus sampling. If nil, uses model default.
	TopP *float64 `json:"top_p,omitempty"`
}

// IsEmpty returns true if the model is not configured.
func (m SelectedModel) IsEmpty() bool {
	return m.Model == ""
}

// String returns a human-readable representation of the model.
func (m SelectedModel) String() string {
	if m.Provider != "" {
		return m.Provider + "/" + m.Model
	}
	return m.Model
}

// Config represents the CLI configuration for ayo.
type Config struct {
	Schema         string           `json:"$schema,omitempty"`
	AgentsDir      string           `json:"agents_dir,omitempty"`
	SystemPrefix   string           `json:"system_prefix,omitempty"`
	SystemSuffix   string           `json:"system_suffix,omitempty"`
	SkillsDir      string           `json:"skills_dir,omitempty"`
	EmbeddingModel string           `json:"embedding_model,omitempty"`

	// Models configures the large and small models with their parameters.
	// This is the preferred way to configure models.
	Models map[ModelType]SelectedModel `json:"models,omitempty"`

	// DefaultModel is the legacy field for the primary model (deprecated, use Models["large"]).
	DefaultModel string `json:"default_model,omitempty"`
	// SmallModel is the legacy field for the small/fast model (deprecated, use Models["small"]).
	SmallModel string `json:"small_model,omitempty"`
	OllamaHost     string           `json:"ollama_host,omitempty"`
	CatwalkBaseURL string           `json:"catwalk_base_url,omitempty"`
	Provider       catwalk.Provider `json:"provider,omitempty"`
	Embedding      EmbeddingConfig  `json:"embedding,omitempty"`

	// Flows configuration
	Flows FlowsConfig `json:"flows,omitempty"`

	// Providers configures the pluggable provider system.
	// This enables swapping implementations for memory, sandbox, embedding, and observers.
	Providers ProvidersConfig `json:"providers,omitempty"`

	// Delegates maps task types to agent handles for global delegation.
	// Example: {"coding": "@crush", "research": "@research"}
	Delegates map[string]string `json:"delegates,omitempty"`

	// DefaultTools maps tool type aliases to concrete tool names.
	// Example: {"search": "searxng"}
	// This allows agents to use generic tool types that resolve to user-configured tools.
	DefaultTools map[string]string `json:"default_tools,omitempty"`
}

// FlowsConfig configures the flows system.
type FlowsConfig struct {
	// HistoryRetentionDays is the maximum age of flow run history in days.
	// Runs older than this are pruned. Default: 30.
	HistoryRetentionDays int `json:"history_retention_days,omitempty"`

	// HistoryMaxRuns is the maximum number of flow runs to keep.
	// Excess runs are pruned (oldest first). Default: 1000.
	HistoryMaxRuns int `json:"history_max_runs,omitempty"`
}

// ProvidersConfig configures the pluggable provider system.
type ProvidersConfig struct {
	// Active maps provider types to the active provider name for that type.
	// Valid types: "memory", "sandbox", "embedding", "observer"
	// Example: {"memory": "zettelkasten", "sandbox": "apple-container"}
	Active map[string]string `json:"active,omitempty"`

	// Memory contains configuration for the active memory provider.
	Memory MemoryProviderConfig `json:"memory,omitempty"`

	// Sandbox contains configuration for the active sandbox provider.
	Sandbox SandboxProviderConfig `json:"sandbox,omitempty"`

	// Embedding contains configuration for the active embedding provider.
	// Note: Legacy EmbeddingConfig is still supported at the top level for compatibility.
	Embedding EmbeddingProviderConfig `json:"embedding,omitempty"`

	// Observer contains configuration for the active observer provider.
	Observer ObserverProviderConfig `json:"observer,omitempty"`
}

// MemoryProviderConfig configures the memory provider.
type MemoryProviderConfig struct {
	// Directory is the base directory for memory storage.
	// Defaults to ~/.local/share/ayo/memory/
	Directory string `json:"directory,omitempty"`

	// IndexPath is the path to the SQLite index (derived, rebuildable).
	// Defaults to {Directory}/.index.sqlite
	IndexPath string `json:"index_path,omitempty"`

	// AutoMerge enables automatic merging of conflicting memories.
	AutoMerge bool `json:"auto_merge,omitempty"`
}

// SandboxProviderConfig configures the sandbox provider.
type SandboxProviderConfig struct {
	// Image is the base container image to use.
	// Defaults to busybox with POSIX tools.
	Image string `json:"image,omitempty"`

	// Pool configures the sandbox pool.
	Pool SandboxPoolConfig `json:"pool,omitempty"`

	// Network configures default network settings.
	Network SandboxNetworkConfig `json:"network,omitempty"`

	// Resources configures default resource limits.
	Resources SandboxResourcesConfig `json:"resources,omitempty"`
}

// SandboxPoolConfig configures the sandbox pool.
type SandboxPoolConfig struct {
	// MinSize is the minimum number of warm sandboxes to maintain.
	// Defaults to 1.
	MinSize int `json:"min_size,omitempty"`

	// MaxSize is the maximum number of sandboxes allowed.
	// Defaults to 4.
	MaxSize int `json:"max_size,omitempty"`

	// IdleTimeout is how long to keep idle sandboxes before recycling.
	// Defaults to 30m. Format: Go duration string (e.g., "30m", "1h").
	IdleTimeout string `json:"idle_timeout,omitempty"`
}

// SandboxNetworkConfig configures sandbox networking.
type SandboxNetworkConfig struct {
	// Enabled determines if network access is allowed. Defaults to true.
	Enabled *bool `json:"enabled,omitempty"`

	// DNS servers to use. Defaults to system DNS.
	DNS []string `json:"dns,omitempty"`
}

// SandboxResourcesConfig configures sandbox resource limits.
type SandboxResourcesConfig struct {
	// CPUs is the number of CPUs to allocate. Defaults to 2.
	CPUs int `json:"cpus,omitempty"`

	// MemoryMB is the memory limit in megabytes. Defaults to 2048.
	MemoryMB int64 `json:"memory_mb,omitempty"`

	// DiskMB is the disk limit in megabytes. Defaults to 10240.
	DiskMB int64 `json:"disk_mb,omitempty"`
}

// EmbeddingProviderConfig configures the embedding provider.
// This is the new providers.embedding config; legacy embedding config
// at the top level is still supported for compatibility.
type EmbeddingProviderConfig struct {
	// Model is the embedding model to use.
	Model string `json:"model,omitempty"`

	// Endpoint is the API endpoint for remote providers.
	Endpoint string `json:"endpoint,omitempty"`

	// Dimensions overrides the default embedding dimensions.
	// Most providers auto-detect this from the model.
	Dimensions int `json:"dimensions,omitempty"`
}

// ObserverProviderConfig configures the observer provider.
type ObserverProviderConfig struct {
	// Enabled determines if the observer is active. Defaults to true.
	Enabled *bool `json:"enabled,omitempty"`

	// BatchSize is the number of messages to process at once.
	// Defaults to 10.
	BatchSize int `json:"batch_size,omitempty"`

	// DebounceMs is how long to wait before processing messages.
	// Defaults to 1000 (1 second).
	DebounceMs int `json:"debounce_ms,omitempty"`
}

// AyoSandboxConfig configures the dedicated @ayo orchestrator sandbox.
// Location: ~/.config/ayo/ayo-sandbox.json
type AyoSandboxConfig struct {
	// Schema enables IDE validation/autocomplete.
	Schema string `json:"$schema,omitempty"`

	// Image is the container image for @ayo's sandbox.
	// Defaults to the standard sandbox image.
	Image string `json:"image,omitempty"`

	// Resources configures CPU, memory, and disk limits for @ayo's sandbox.
	Resources SandboxResourcesConfig `json:"resources,omitempty"`

	// Packages lists packages to install in @ayo's sandbox during setup.
	// These are installed via the sandbox's package manager (e.g., apk, apt).
	Packages []string `json:"packages,omitempty"`

	// Mounts specifies additional paths to mount into @ayo's sandbox.
	// Format: "host_path:container_path" or just "path" for same path on both.
	Mounts []string `json:"mounts,omitempty"`

	// Network enables network access in @ayo's sandbox.
	// Defaults to true.
	Network *bool `json:"network,omitempty"`
}

// SquadConfig configures a squad sandbox for agent teams.
// Location: ~/.config/ayo/squads/{name}.json
type SquadConfig struct {
	// Schema enables IDE validation/autocomplete.
	Schema string `json:"$schema,omitempty"`

	// Name is the squad identifier (derived from filename if not specified).
	Name string `json:"name,omitempty"`

	// Description provides a human-readable description of the squad's purpose.
	Description string `json:"description,omitempty"`

	// Image is the container image for the squad sandbox.
	// Defaults to the standard sandbox image.
	Image string `json:"image,omitempty"`

	// Resources configures CPU, memory, and disk limits for the squad sandbox.
	Resources SandboxResourcesConfig `json:"resources,omitempty"`

	// Packages lists packages to install in the squad sandbox during setup.
	Packages []string `json:"packages,omitempty"`

	// Mounts specifies additional paths to mount into the squad sandbox.
	// Format: "host_path:container_path" or just "path" for same path on both.
	Mounts []string `json:"mounts,omitempty"`

	// Network enables network access in the squad sandbox.
	// Defaults to true.
	Network *bool `json:"network,omitempty"`

	// Ephemeral indicates whether the squad sandbox is destroyed after the session.
	// Defaults to false (persistent).
	Ephemeral bool `json:"ephemeral,omitempty"`

	// Agents lists agent handles that are members of this squad.
	// Example: ["@frontend", "@backend", "@qa"]
	Agents []string `json:"agents,omitempty"`

	// WorkspaceMount is the host path to mount as /workspace in the squad sandbox.
	// This is where agents work on code.
	WorkspaceMount string `json:"workspace_mount,omitempty"`

	// OutputPath is the host path where work products are synced after completion.
	OutputPath string `json:"output_path,omitempty"`
}

// DefaultProvidersConfig returns the default provider configuration.
func DefaultProvidersConfig() ProvidersConfig {
	networkEnabled := true
	observerEnabled := true

	return ProvidersConfig{
		Active: map[string]string{
			"memory":    "zettelkasten",
			"sandbox":   "none", // No sandbox by default until implemented
			"embedding": "ollama",
			"observer":  "memory-extractor",
		},
		Memory: MemoryProviderConfig{
			AutoMerge: true,
		},
		Sandbox: SandboxProviderConfig{
			Image: "busybox:latest",
			Pool: SandboxPoolConfig{
				MinSize:     1,
				MaxSize:     4,
				IdleTimeout: "30m",
			},
			Network: SandboxNetworkConfig{
				Enabled: &networkEnabled,
			},
			Resources: SandboxResourcesConfig{
				CPUs:     2,
				MemoryMB: 2048,
				DiskMB:   10240,
			},
		},
		Embedding: EmbeddingProviderConfig{
			Model: "nomic-embed-text",
		},
		Observer: ObserverProviderConfig{
			Enabled:    &observerEnabled,
			BatchSize:  10,
			DebounceMs: 1000,
		},
	}
}

// EmbeddingConfig configures the embedding system.
type EmbeddingConfig struct {
	// Provider is the embedding provider. Use "local" for offline embeddings (default),
	// or "openai", "voyage", "ollama" for cloud-based embeddings.
	Provider string `json:"provider,omitempty"`

	// Model is the embedding model to use (provider-specific).
	Model string `json:"model,omitempty"`

	// APIKey is the API key for cloud providers (can also use environment variables).
	APIKey string `json:"api_key,omitempty"`

	// Endpoint overrides the default API endpoint for the provider.
	Endpoint string `json:"endpoint,omitempty"`
}

func defaultCatwalkURL() string {
	if env := strings.TrimSpace(os.Getenv("CATWALK_URL")); env != "" {
		return env
	}
	return "http://localhost:8080"
}

// Default returns a Config populated with default values.
func Default() Config {
	// Get the best default models based on available credentials
	defaultModel := GetDefaultModelForConfiguredProvider()
	smallModel := GetDefaultSmallModelForConfiguredProvider()

	return Config{
		AgentsDir:      paths.AgentsDir(),
		SystemPrefix:   "", // Uses paths.FindPromptFile("system-prefix.md")
		SystemSuffix:   "", // Uses paths.FindPromptFile("system-suffix.md")
		SkillsDir:      paths.SkillsDir(),
		DefaultModel:   defaultModel,
		SmallModel:     smallModel,
		EmbeddingModel: "ollama/nomic-embed-text", // Ollama-only for embeddings (local)
		OllamaHost:     "http://localhost:11434",
		CatwalkBaseURL: defaultCatwalkURL(),
		Provider: catwalk.Provider{
			Name:        "openai",
			ID:          catwalk.InferenceProviderOpenAI,
			Type:        catwalk.TypeOpenAI,
			APIEndpoint: "https://api.openai.com/v1",
		},
		Embedding: EmbeddingConfig{
			Provider: "ollama",
			Model:    "nomic-embed-text",
		},
		Flows: FlowsConfig{
			HistoryRetentionDays: 30,
			HistoryMaxRuns:       1000,
		},
		Providers: DefaultProvidersConfig(),
	}
}

// Load reads configuration from the given path, falling back to defaults when missing.
func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if strings.TrimSpace(cfg.CatwalkBaseURL) == "" {
		cfg.CatwalkBaseURL = defaultCatwalkURL()
	}

	return cfg, nil
}

// DefaultPath returns the default config file path.
func DefaultPath() string {
	return paths.ConfigFile()
}

// Save writes the configuration to the given path.
func Save(path string, cfg Config) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// SetDelegate sets a delegate in the global config.
// Returns the previous value if any.
func SetDelegate(taskType, agentHandle string) (previous string, err error) {
	cfgPath := DefaultPath()
	cfg, err := Load(cfgPath)
	if err != nil {
		return "", err
	}

	if cfg.Delegates == nil {
		cfg.Delegates = make(map[string]string)
	}

	previous = cfg.Delegates[taskType]
	cfg.Delegates[taskType] = agentHandle

	if err := Save(cfgPath, cfg); err != nil {
		return previous, err
	}

	return previous, nil
}

// GetDelegate returns the current delegate for a task type from global config.
func GetDelegate(taskType string) (string, error) {
	cfg, err := Load(DefaultPath())
	if err != nil {
		return "", err
	}

	if cfg.Delegates == nil {
		return "", nil
	}

	return cfg.Delegates[taskType], nil
}

// SetDefaultTool sets a default tool mapping in the global config.
// Returns the previous value if any.
func SetDefaultTool(toolType, toolName string) (previous string, err error) {
	cfgPath := DefaultPath()
	cfg, err := Load(cfgPath)
	if err != nil {
		return "", err
	}

	if cfg.DefaultTools == nil {
		cfg.DefaultTools = make(map[string]string)
	}

	previous = cfg.DefaultTools[toolType]
	cfg.DefaultTools[toolType] = toolName

	if err := Save(cfgPath, cfg); err != nil {
		return previous, err
	}

	return previous, nil
}

// GetDefaultTool returns the current default tool for a tool type from global config.
func GetDefaultTool(toolType string) (string, error) {
	cfg, err := Load(DefaultPath())
	if err != nil {
		return "", err
	}

	if cfg.DefaultTools == nil {
		return "", nil
	}

	return cfg.DefaultTools[toolType], nil
}

// AyoSandboxConfigPath returns the path to the @ayo sandbox config file.
func AyoSandboxConfigPath() string {
	return filepath.Join(paths.ConfigDir(), "ayo-sandbox.json")
}

// DefaultAyoSandboxConfig returns the default @ayo sandbox configuration.
func DefaultAyoSandboxConfig() AyoSandboxConfig {
	networkEnabled := true
	return AyoSandboxConfig{
		Image: "busybox:latest",
		Resources: SandboxResourcesConfig{
			CPUs:     2,
			MemoryMB: 2048,
			DiskMB:   10240,
		},
		Packages: []string{},
		Mounts:   []string{},
		Network:  &networkEnabled,
	}
}

// LoadAyoSandboxConfig reads the @ayo sandbox configuration.
// Returns defaults if the config file doesn't exist.
func LoadAyoSandboxConfig() (AyoSandboxConfig, error) {
	cfg := DefaultAyoSandboxConfig()

	data, err := os.ReadFile(AyoSandboxConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read ayo sandbox config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse ayo sandbox config: %w", err)
	}

	return cfg, nil
}

// SaveAyoSandboxConfig writes the @ayo sandbox configuration.
func SaveAyoSandboxConfig(cfg AyoSandboxConfig) error {
	path := AyoSandboxConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create ayo sandbox config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ayo sandbox config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write ayo sandbox config: %w", err)
	}

	return nil
}

// SquadsConfigDir returns the directory for squad config files.
func SquadsConfigDir() string {
	return filepath.Join(paths.ConfigDir(), "squads")
}

// SquadConfigPath returns the path to a squad's config file.
func SquadConfigPath(name string) string {
	return filepath.Join(SquadsConfigDir(), name+".json")
}

// DefaultSquadConfig returns the default squad configuration.
func DefaultSquadConfig(name string) SquadConfig {
	networkEnabled := true
	return SquadConfig{
		Name: name,
		Resources: SandboxResourcesConfig{
			CPUs:     2,
			MemoryMB: 2048,
			DiskMB:   10240,
		},
		Packages:  []string{},
		Mounts:    []string{},
		Network:   &networkEnabled,
		Ephemeral: false,
		Agents:    []string{},
	}
}

// LoadSquadConfig reads a squad configuration by name.
// Returns defaults if the config file doesn't exist.
func LoadSquadConfig(name string) (SquadConfig, error) {
	cfg := DefaultSquadConfig(name)

	data, err := os.ReadFile(SquadConfigPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read squad config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse squad config: %w", err)
	}

	// Ensure name is set
	if cfg.Name == "" {
		cfg.Name = name
	}

	return cfg, nil
}

// SaveSquadConfig writes a squad configuration.
func SaveSquadConfig(cfg SquadConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("squad name is required")
	}

	path := SquadConfigPath(cfg.Name)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create squads config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal squad config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write squad config: %w", err)
	}

	return nil
}

// ListSquadConfigs returns the names of all configured squads.
func ListSquadConfigs() ([]string, error) {
	dir := SquadsConfigDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read squads directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			names = append(names, strings.TrimSuffix(name, ".json"))
		}
	}
	return names, nil
}

// DeleteSquadConfig removes a squad configuration file.
func DeleteSquadConfig(name string) error {
	path := SquadConfigPath(name)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete squad config: %w", err)
	}
	return nil
}
