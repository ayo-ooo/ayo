package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// CreateProject creates a minimal valid project structure for testing.
// Returns the path to the project directory.
func CreateProject(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	projectDir := filepath.Join(dir, name)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("creating project directory: %v", err)
	}

	config := `[agent]
name = "` + name + `"
version = "1.0.0"
description = "Test agent"

[model]
requires_structured_output = false
requires_tools = false
requires_vision = false
suggested = ["claude-sonnet-4-6"]
default = "claude-sonnet-4-6"

[defaults]
temperature = 0.7
max_tokens = 4096
`

	if err := os.WriteFile(filepath.Join(projectDir, "config.toml"), []byte(config), 0644); err != nil {
		t.Fatalf("writing config.toml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(projectDir, "system.md"), []byte("# System\n\nYou are a helpful assistant."), 0644); err != nil {
		t.Fatalf("writing system.md: %v", err)
	}

	return projectDir
}

// CreateProjectWithFiles creates a project with additional files.
func CreateProjectWithFiles(t *testing.T, name string, files map[string]string) string {
	t.Helper()
	projectDir := CreateProject(t, name)

	for filename, content := range files {
		path := filepath.Join(projectDir, filename)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("creating directory for %s: %v", filename, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("writing %s: %v", filename, err)
		}
	}

	return projectDir
}

// ValidInputSchema returns a valid input JSON schema for testing.
func ValidInputSchema() string {
	return `{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "Search query"
			}
		},
		"required": ["query"]
	}`
}

// ValidOutputSchema returns a valid output JSON schema for testing.
func ValidOutputSchema() string {
	return `{
		"type": "object",
		"properties": {
			"result": {
				"type": "string",
				"description": "The result"
			}
		}
	}`
}

// SchemaWithDefaults returns a schema with default values for testing.
func SchemaWithDefaults() string {
	return `{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"default": "default-name"
			},
			"count": {
				"type": "integer",
				"default": 10
			},
			"enabled": {
				"type": "boolean",
				"default": true
			},
			"ratio": {
				"type": "number",
				"default": 0.5
			}
		}
	}`
}

// SchemaWithCLIExtensions returns a schema with CLI extensions for testing.
func SchemaWithCLIExtensions() string {
	return `{
		"type": "object",
		"properties": {
			"input": {
				"type": "string",
				"x-cli-position": 1
			},
			"output": {
				"type": "string",
				"x-cli-flag": "--out",
				"x-cli-short": "o"
			},
			"file": {
				"type": "string",
				"x-cli-file": true
			}
		}
	}`
}

// InvalidJSON returns invalid JSON for error testing.
func InvalidJSON() string {
	return `{invalid json}`
}

// MinimalConfig returns a minimal valid TOML config.
func MinimalConfig(name string) string {
	return `[agent]
name = "` + name + `"
version = "1.0.0"
description = "Test agent"

[model]
default = "claude-sonnet-4-6"

[defaults]
temperature = 0.7
max_tokens = 4096
`
}
