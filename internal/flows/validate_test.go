package flows

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateInput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a packaged flow with schema
	pkgDir := filepath.Join(tmpDir, "schema-flow")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: schema-flow
# description: Flow with schema

echo "$1"
`
	if err := os.WriteFile(filepath.Join(pkgDir, "flow.sh"), []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	inputSchema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "name": {"type": "string"},
    "count": {"type": "integer"}
  },
  "required": ["name"]
}
`
	if err := os.WriteFile(filepath.Join(pkgDir, "input.jsonschema"), []byte(inputSchema), 0644); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(pkgDir)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid input",
			input:   `{"name": "test", "count": 42}`,
			wantErr: false,
		},
		{
			name:    "valid input minimal",
			input:   `{"name": "test"}`,
			wantErr: false,
		},
		{
			name:    "missing required field",
			input:   `{"count": 42}`,
			wantErr: true,
		},
		{
			name:    "wrong type",
			input:   `{"name": 123}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInput(flow, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateInput_NoSchema(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: no-schema-flow
# description: Flow without schema

echo "$1"
`
	flowPath := filepath.Join(tmpDir, "no-schema.sh")
	if err := os.WriteFile(flowPath, []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(flowPath)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	// Should pass any input when no schema
	err = ValidateInput(flow, `{"anything": "goes"}`)
	if err != nil {
		t.Errorf("ValidateInput() unexpected error: %v", err)
	}
}

func TestValidateOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a packaged flow with output schema
	pkgDir := filepath.Join(tmpDir, "output-flow")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: output-flow
# description: Flow with output schema

echo "$1"
`
	if err := os.WriteFile(filepath.Join(pkgDir, "flow.sh"), []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	outputSchema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "result": {"type": "string"},
    "success": {"type": "boolean"}
  },
  "required": ["result", "success"]
}
`
	if err := os.WriteFile(filepath.Join(pkgDir, "output.jsonschema"), []byte(outputSchema), 0644); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(pkgDir)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	tests := []struct {
		name         string
		output       string
		wantWarnings bool
	}{
		{
			name:         "valid output",
			output:       `{"result": "done", "success": true}`,
			wantWarnings: false,
		},
		{
			name:         "missing field",
			output:       `{"result": "done"}`,
			wantWarnings: true,
		},
		{
			name:         "wrong type",
			output:       `{"result": 123, "success": true}`,
			wantWarnings: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := ValidateOutput(flow, tt.output)
			if (len(warnings) > 0) != tt.wantWarnings {
				t.Errorf("ValidateOutput() warnings = %v, wantWarnings %v", warnings, tt.wantWarnings)
			}
		})
	}
}

func TestSchemaValidationError(t *testing.T) {
	err := &SchemaValidationError{
		Message: "validation failed",
		Details: []string{"missing field 'name'", "wrong type for 'count'"},
	}

	expected := "validation failed: [missing field 'name' wrong type for 'count']"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// Empty details
	err2 := &SchemaValidationError{
		Message: "simple error",
		Details: nil,
	}
	if err2.Error() != "simple error" {
		t.Errorf("Error() = %q, want %q", err2.Error(), "simple error")
	}
}
