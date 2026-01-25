package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexcabrera/ayo/internal/paths"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// Config represents the CLI configuration for ayo.
type Config struct {
	Schema         string           `json:"$schema,omitempty"`
	AgentsDir      string           `json:"agents_dir,omitempty"`
	SystemPrefix   string           `json:"system_prefix,omitempty"`
	SystemSuffix   string           `json:"system_suffix,omitempty"`
	SkillsDir      string           `json:"skills_dir,omitempty"`
	DefaultModel   string           `json:"default_model,omitempty"`
	SmallModel     string           `json:"small_model,omitempty"`
	EmbeddingModel string           `json:"embedding_model,omitempty"`
	OllamaHost     string           `json:"ollama_host,omitempty"`
	CatwalkBaseURL string           `json:"catwalk_base_url,omitempty"`
	Provider       catwalk.Provider `json:"provider,omitempty"`
	Embedding      EmbeddingConfig  `json:"embedding,omitempty"`

	// Delegates maps task types to agent handles for global delegation.
	// Example: {"coding": "@crush", "research": "@research"}
	Delegates map[string]string `json:"delegates,omitempty"`

	// DefaultTools maps tool type aliases to concrete tool names.
	// Example: {"search": "searxng"}
	// This allows agents to use generic tool types that resolve to user-configured tools.
	DefaultTools map[string]string `json:"default_tools,omitempty"`
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

func apiKeyEnvForProvider(p catwalk.Provider) string {
	if p.ID == "" {
		return ""
	}
	return strings.ToUpper(string(p.ID)) + "_API_KEY"
}

func defaultCatwalkURL() string {
	if env := strings.TrimSpace(os.Getenv("CATWALK_URL")); env != "" {
		return env
	}
	return "http://localhost:8080"
}

// Default returns a Config populated with default values.
func Default() Config {
	return Config{
		AgentsDir:      paths.AgentsDir(),
		SystemPrefix:   "", // Uses paths.FindPromptFile("system-prefix.md")
		SystemSuffix:   "", // Uses paths.FindPromptFile("system-suffix.md")
		SkillsDir:      paths.SkillsDir(),
		DefaultModel:   "gpt-4.1",
		SmallModel:     "ollama/ministral-3:3b",
		EmbeddingModel: "ollama/nomic-embed-text",
		OllamaHost:     "http://localhost:11434",
		CatwalkBaseURL: defaultCatwalkURL(),
		Provider: catwalk.Provider{
			Name:        "openai",
			ID:          catwalk.InferenceProviderOpenAI,
			APIEndpoint: "https://api.openai.com/v1",
		},
		Embedding: EmbeddingConfig{
			Provider: "ollama",
			Model:    "nomic-embed-text",
		},
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
