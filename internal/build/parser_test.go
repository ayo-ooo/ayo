package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/build/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteConfig tests the WriteConfig function
func TestWriteConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-write-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(config, configPath)
	require.NoError(t, err)

	// Verify the file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Read and verify the content
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "test-agent")
}

// TestWriteConfigError tests error handling in WriteConfig
func TestWriteConfigError(t *testing.T) {
	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	// Try to write to a non-existent directory
	invalidPath := "/non/existent/path/config.toml"
	err := WriteConfig(config, invalidPath)
	assert.Error(t, err)
}

// TestParseConfigBasic tests basic config parsing
func TestParseConfigBasic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-parse-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
			Model:       "gpt-4-turbo",
		},
		CLI: types.CLIConfig{
			Mode:        "freeform",
			Description: "Test CLI",
		},
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(config, configPath)
	require.NoError(t, err)

	// Parse the config back
	parsedConfig, err := ParseConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "test-agent", parsedConfig.Agent.Name)
	assert.Equal(t, "Test agent", parsedConfig.Agent.Description)
	assert.Equal(t, "gpt-4-turbo", parsedConfig.Agent.Model)
	assert.Equal(t, "freeform", parsedConfig.CLI.Mode)
}

// TestParseConfigWithInvalidTOML tests parsing invalid TOML
func TestParseConfigWithInvalidTOML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-invalid-toml-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.toml")
	invalidTOML := "this is not valid TOML [[["
	err = os.WriteFile(configPath, []byte(invalidTOML), 0644)
	require.NoError(t, err)

	_, err = ParseConfig(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse TOML")
}

// TestParseConfigWithInvalidModel tests parsing with invalid model
func TestParseConfigWithInvalidModel(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-invalid-model-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
			Model:       "invalid-model-name",
		},
		CLI: types.CLIConfig{
			Mode:        "freeform",
			Description: "Test CLI",
		},
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(config, configPath)
	require.NoError(t, err)

	_, err = ParseConfig(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported model")
}

// TestParseConfigWithInvalidSchema tests parsing with invalid JSON schema
func TestParseConfigWithInvalidSchema(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-invalid-schema-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
			Model:       "gpt-4-turbo",
		},
		CLI: types.CLIConfig{
			Mode:        "freeform",
			Description: "Test CLI",
		},
		Input: types.InputConfig{
			Schema: map[string]any{
				"type": "object",
				"properties": "not an object",
			},
		},
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(config, configPath)
	require.NoError(t, err)

	_, err = ParseConfig(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input schema validation")
}

// TestValidateJSONSchemaValid tests validating a valid JSON schema
func TestValidateJSONSchemaValid(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"name"},
	}

	err := validateJSONSchema(schema, "test")
	assert.NoError(t, err)
}

// TestValidateJSONSchemaInvalid tests validating an invalid JSON schema
func TestValidateJSONSchemaInvalid(t *testing.T) {
	// Invalid schema: object properties must be objects
	schema := map[string]any{
		"type": "object",
		"properties": "not an object",
	}

	err := validateJSONSchema(schema, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compile schema")
}

// TestValidateJSONSchemaUnserializable tests schema that can't be serialized
func TestValidateJSONSchemaUnserializable(t *testing.T) {
	// Schema with a value that can't be serialized to JSON
	schema := map[string]any{
		"type": func() {},
	}

	err := validateJSONSchema(schema, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serialize schema")
}

// TestLoadConfigFromDirConfigToml tests loading config.toml from directory
func TestLoadConfigFromDirConfigToml(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-load-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode:        "freeform",
			Description: "Test CLI",
		},
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(config, configPath)
	require.NoError(t, err)

	loadedConfig, loadedPath, err := LoadConfigFromDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "test-agent", loadedConfig.Agent.Name)
	assert.Equal(t, configPath, loadedPath)
}

// TestLoadConfigFromDirNotFound tests loading from directory without config
func TestLoadConfigFromDirNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-no-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	_, _, err = LoadConfigFromDir(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no config.toml or team.toml found")
}

// TestLoadConfigFromDirInvalidConfig tests loading with invalid config file
func TestLoadConfigFromDirInvalidConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-invalid-config-dir-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.toml")
	invalidTOML := "invalid toml content"
	err = os.WriteFile(configPath, []byte(invalidTOML), 0644)
	require.NoError(t, err)

	_, _, err = LoadConfigFromDir(tmpDir)
	assert.Error(t, err)
}



// TestParseConfigWithBothSchemas tests parsing with both input and output schemas
func TestParseConfigWithBothSchemas(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-both-schemas-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode:        "freeform",
			Description: "Test CLI",
		},
		Input: types.InputConfig{
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"prompt": map[string]any{
						"type": "string",
					},
				},
			},
		},
		Output: types.OutputConfig{
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{
						"type": "string",
					},
				},
			},
		},
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(config, configPath)
	require.NoError(t, err)

	parsedConfig, err := ParseConfig(configPath)
	require.NoError(t, err)
	assert.NotNil(t, parsedConfig.Input.Schema)
	assert.NotNil(t, parsedConfig.Output.Schema)
}

// TestParseConfigWithOutputSchemaOnly tests parsing with only output schema
func TestParseConfigWithOutputSchemaOnly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-output-schema-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent",
		},
		CLI: types.CLIConfig{
			Mode:        "freeform",
			Description: "Test CLI",
		},
		Output: types.OutputConfig{
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{
						"type": "string",
					},
				},
			},
		},
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(config, configPath)
	require.NoError(t, err)

	parsedConfig, err := ParseConfig(configPath)
	require.NoError(t, err)
	assert.Nil(t, parsedConfig.Input.Schema)
	assert.NotNil(t, parsedConfig.Output.Schema)
}
