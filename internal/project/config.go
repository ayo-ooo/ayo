package project

import (
	"fmt"
	"os"

	toml "github.com/BurntSushi/toml"
)

type configToml struct {
	Agent    agentSection    `toml:"agent"`
	Model    modelSection    `toml:"model"`
	Defaults defaultsSection `toml:"defaults"`
}

type agentSection struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
}

type modelSection struct {
	RequiresStructuredOutput bool     `toml:"requires_structured_output"`
	RequiresTools            bool     `toml:"requires_tools"`
	RequiresVision           bool     `toml:"requires_vision"`
	Suggested                []string `toml:"suggested"`
	Default                  string   `toml:"default"`
}

type defaultsSection struct {
	Temperature float64 `toml:"temperature"`
	MaxTokens   int     `toml:"max_tokens"`
}

func ParseConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg configToml
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &AgentConfig{
		Name:        cfg.Agent.Name,
		Version:     cfg.Agent.Version,
		Description: cfg.Agent.Description,
		Model: ModelRequirements{
			RequiresStructuredOutput: cfg.Model.RequiresStructuredOutput,
			RequiresTools:            cfg.Model.RequiresTools,
			RequiresVision:           cfg.Model.RequiresVision,
			Suggested:                cfg.Model.Suggested,
			Default:                  cfg.Model.Default,
		},
		Defaults: AgentDefaults{
			Temperature: cfg.Defaults.Temperature,
			MaxTokens:   cfg.Defaults.MaxTokens,
		},
	}, nil
}
