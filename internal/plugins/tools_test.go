package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadToolDefinition(t *testing.T) {
	dir := t.TempDir()

	// Create tools directory
	toolDir := filepath.Join(dir, "tools", "my-tool")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create valid tool definition
	toolJSON := `{
		"name": "my-tool",
		"description": "A test tool",
		"command": "echo",
		"args": ["hello"],
		"parameters": [
			{
				"name": "message",
				"description": "The message to echo",
				"type": "string",
				"required": true
			}
		],
		"timeout": 30
	}`

	if err := os.WriteFile(filepath.Join(toolDir, "tool.json"), []byte(toolJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	// Load tool definition
	td, err := LoadToolDefinition(dir, "my-tool")
	if err != nil {
		t.Fatalf("LoadToolDefinition failed: %v", err)
	}

	if td.Name != "my-tool" {
		t.Errorf("Name = %q, want %q", td.Name, "my-tool")
	}
	if td.Command != "echo" {
		t.Errorf("Command = %q, want %q", td.Command, "echo")
	}
	if len(td.Parameters) != 1 {
		t.Errorf("Parameters count = %d, want 1", len(td.Parameters))
	}
	if td.Timeout != 30 {
		t.Errorf("Timeout = %d, want 30", td.Timeout)
	}
}

func TestToolDefinitionValidation(t *testing.T) {
	tests := []struct {
		name    string
		td      ToolDefinition
		wantErr bool
	}{
		{
			name:    "missing name",
			td:      ToolDefinition{Description: "test", Command: "echo"},
			wantErr: true,
		},
		{
			name:    "missing description",
			td:      ToolDefinition{Name: "test", Command: "echo"},
			wantErr: true,
		},
		{
			name:    "missing command",
			td:      ToolDefinition{Name: "test", Description: "test"},
			wantErr: true,
		},
		{
			name: "valid minimal",
			td:   ToolDefinition{Name: "test", Description: "test", Command: "echo"},
		},
		{
			name: "valid with parameters",
			td: ToolDefinition{
				Name:        "test",
				Description: "test",
				Command:     "echo",
				Parameters: []ToolParameter{
					{Name: "arg", Description: "An arg", Type: "string"},
				},
			},
		},
		{
			name: "invalid parameter type",
			td: ToolDefinition{
				Name:        "test",
				Description: "test",
				Command:     "echo",
				Parameters: []ToolParameter{
					{Name: "arg", Description: "An arg", Type: "invalid"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.td.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToolDefinitionToJSONSchema(t *testing.T) {
	td := ToolDefinition{
		Name:        "test",
		Description: "test tool",
		Command:     "echo",
		Parameters: []ToolParameter{
			{Name: "required_arg", Description: "Required", Type: "string", Required: true},
			{Name: "optional_arg", Description: "Optional", Type: "number"},
		},
	}

	schema := td.ToJSONSchema()

	// Check type
	if schema["type"] != "object" {
		t.Errorf("schema type = %v, want object", schema["type"])
	}

	// Check required
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required field is not []string")
	}
	if len(required) != 1 || required[0] != "required_arg" {
		t.Errorf("required = %v, want [required_arg]", required)
	}

	// Check properties
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties field is not map")
	}
	if len(props) != 2 {
		t.Errorf("properties count = %d, want 2", len(props))
	}
}

func TestGetRequiredParams(t *testing.T) {
	td := ToolDefinition{
		Name:        "test",
		Description: "test",
		Command:     "echo",
		Parameters: []ToolParameter{
			{Name: "req1", Description: "Required 1", Type: "string", Required: true},
			{Name: "opt1", Description: "Optional", Type: "string"},
			{Name: "req2", Description: "Required 2", Type: "string", Required: true},
		},
	}

	required := td.GetRequiredParams()
	if len(required) != 2 {
		t.Errorf("GetRequiredParams() returned %d, want 2", len(required))
	}
}

func TestGetParamByName(t *testing.T) {
	td := ToolDefinition{
		Name:        "test",
		Description: "test",
		Command:     "echo",
		Parameters: []ToolParameter{
			{Name: "arg1", Description: "Arg 1", Type: "string"},
			{Name: "arg2", Description: "Arg 2", Type: "number"},
		},
	}

	param := td.GetParamByName("arg2")
	if param == nil {
		t.Fatal("GetParamByName returned nil")
	}
	if param.Type != "number" {
		t.Errorf("param.Type = %q, want number", param.Type)
	}

	// Non-existent
	if td.GetParamByName("nonexistent") != nil {
		t.Error("GetParamByName should return nil for non-existent param")
	}
}
