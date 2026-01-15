package config

import (
	"encoding/json"
	"fmt"
	"os"
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
	CatwalkBaseURL string           `json:"catwalk_base_url,omitempty"`
	Provider       catwalk.Provider `json:"provider,omitempty"`
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
		CatwalkBaseURL: defaultCatwalkURL(),
		Provider: catwalk.Provider{
			Name:        "openai",
			ID:          catwalk.InferenceProviderOpenAI,
			APIEndpoint: "https://api.openai.com/v1",
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
