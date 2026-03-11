package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kaptinlin/jsonschema"
	"github.com/pelletier/go-toml/v2"
	"github.com/alexcabrera/ayo/internal/build/types"
)

// WriteConfig writes a config to a TOML file
func WriteConfig(config types.Config, path string) error {
	// Marshal to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal TOML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// ParseConfig reads and parses a config.toml file
func ParseConfig(path string) (*types.Config, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// Parse TOML
	var config types.Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse TOML: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// Validate model name
	if !types.IsValidModel(config.Agent.Model) {
		return nil, fmt.Errorf("unsupported model '%s': must be gpt-*, claude-*, o1-*, or gemini-*", config.Agent.Model)
	}

	// Validate JSON schemas if present
	if config.Input.Schema != nil {
		if err := validateJSONSchema(config.Input.Schema, "input"); err != nil {
			return nil, fmt.Errorf("input schema validation: %w", err)
		}
	}

	if config.Output.Schema != nil {
		if err := validateJSONSchema(config.Output.Schema, "output"); err != nil {
			return nil, fmt.Errorf("output schema validation: %w", err)
		}
	}

	return &config, nil
}

// validateJSONSchema validates a JSON schema (provided as a map[string]any)
func validateJSONSchema(schema map[string]any, context string) error {
	// Convert map to JSON
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("serialize schema: %w", err)
	}

	// Try to compile the schema
	compiler := jsonschema.NewCompiler()
	_, err = compiler.Compile(schemaJSON)
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}

	return nil
}

// LoadConfigFromDir finds and loads config.toml from a directory
func LoadConfigFromDir(dir string) (*types.Config, string, error) {
	// Try config.toml first
	configPath := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		config, err := ParseConfig(configPath)
		if err != nil {
			return nil, "", err
		}
		return config, configPath, nil
	}

	// Try team.toml for teams
	teamPath := filepath.Join(dir, "team.toml")
	if _, err := os.Stat(teamPath); err == nil {
		config, err := ParseConfig(teamPath)
		if err != nil {
			return nil, "", err
		}
		return config, teamPath, nil
	}

	return nil, "", fmt.Errorf("no config.toml or team.toml found in %s", dir)
}
