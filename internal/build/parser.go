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

	// Validate required fields
	if err := validateConfig(&config); err != nil {
		return nil, err
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

	// Validate CLI configuration
	if err := validateCLIConfig(&config.CLI); err != nil {
		return nil, fmt.Errorf("CLI config validation: %w", err)
	}

	return &config, nil
}

// validateConfig checks that required configuration fields are present
func validateConfig(config *types.Config) error {
	// Check agent section
	if config.Agent.Name == "" {
		return fmt.Errorf("agent.name is required")
	}
	if config.Agent.Description == "" {
		return fmt.Errorf("agent.description is required")
	}
	if config.Agent.Model == "" {
		return fmt.Errorf("agent.model is required")
	}

	// Check CLI section
	if config.CLI.Mode == "" {
		return fmt.Errorf("cli.mode is required")
	}
	if config.CLI.Mode != "structured" && config.CLI.Mode != "freeform" && config.CLI.Mode != "hybrid" {
		return fmt.Errorf("cli.mode must be 'structured', 'freeform', or 'hybrid', got '%s'", config.CLI.Mode)
	}

	return nil
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

// validateCLIConfig validates the CLI configuration
func validateCLIConfig(cli *types.CLIConfig) error {
	// Check for duplicate flag names
	flagNames := make(map[string]string)
	for name, flag := range cli.Flags {
		// Check that flag name matches the map key
		if flag.Name != name {
			return fmt.Errorf("flag name mismatch: map key '%s' does not match flag.name '%s'", name, flag.Name)
		}

		// Check flag type
		validTypes := map[string]bool{
			"string": true,
			"int":    true,
			"float":  true,
			"bool":   true,
			"array":  true,
		}
		if !validTypes[flag.Type] {
			return fmt.Errorf("invalid flag type '%s' for flag '%s' (must be string, int, float, bool, or array)", flag.Type, flag.Name)
		}

		// Check for duplicate short flags
		if flag.Short != "" {
			if existing, exists := flagNames[flag.Short]; exists {
				return fmt.Errorf("duplicate short flag '-%s' used by both '%s' and '%s'", flag.Short, existing, flag.Name)
			}
			flagNames[flag.Short] = flag.Name
		}

		// Check position is non-negative
		if flag.Position < 0 {
			return fmt.Errorf("invalid position %d for flag '%s' (must be >= 0)", flag.Position, flag.Name)
		}
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
