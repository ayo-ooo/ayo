package flows

import (
	"testing"
)

func TestFlowSource(t *testing.T) {
	tests := []struct {
		source FlowSource
		want   string
	}{
		{FlowSourceBuiltin, "built-in"},
		{FlowSourceUser, "user"},
		{FlowSourceProject, "project"},
	}

	for _, tt := range tests {
		if string(tt.source) != tt.want {
			t.Errorf("FlowSource = %q, want %q", tt.source, tt.want)
		}
	}
}

func TestFlow_HasSchemas(t *testing.T) {
	t.Run("no schemas", func(t *testing.T) {
		f := &Flow{}
		if f.HasInputSchema() {
			t.Error("expected HasInputSchema() = false")
		}
		if f.HasOutputSchema() {
			t.Error("expected HasOutputSchema() = false")
		}
	})

	t.Run("with input schema", func(t *testing.T) {
		f := &Flow{InputSchemaPath: "/path/to/input.jsonschema"}
		if !f.HasInputSchema() {
			t.Error("expected HasInputSchema() = true")
		}
		if f.HasOutputSchema() {
			t.Error("expected HasOutputSchema() = false")
		}
	})

	t.Run("with output schema", func(t *testing.T) {
		f := &Flow{OutputSchemaPath: "/path/to/output.jsonschema"}
		if f.HasInputSchema() {
			t.Error("expected HasInputSchema() = false")
		}
		if !f.HasOutputSchema() {
			t.Error("expected HasOutputSchema() = true")
		}
	})

	t.Run("with both schemas", func(t *testing.T) {
		f := &Flow{
			InputSchemaPath:  "/path/to/input.jsonschema",
			OutputSchemaPath: "/path/to/output.jsonschema",
		}
		if !f.HasInputSchema() {
			t.Error("expected HasInputSchema() = true")
		}
		if !f.HasOutputSchema() {
			t.Error("expected HasOutputSchema() = true")
		}
	})
}

func TestFlowConstruction(t *testing.T) {
	f := &Flow{
		Name:        "test-flow",
		Description: "A test flow",
		Path:        "/path/to/test-flow.sh",
		Dir:         "/path/to",
		Source:      FlowSourceUser,
		Metadata: FlowMetadata{
			Version: "1.0.0",
			Author:  "test",
		},
		Raw: FlowRaw{
			Frontmatter: map[string]string{
				"name":        "test-flow",
				"description": "A test flow",
			},
			Script: "echo hello",
		},
	}

	if f.Name != "test-flow" {
		t.Errorf("Name = %q, want %q", f.Name, "test-flow")
	}
	if f.Description != "A test flow" {
		t.Errorf("Description = %q, want %q", f.Description, "A test flow")
	}
	if f.Source != FlowSourceUser {
		t.Errorf("Source = %q, want %q", f.Source, FlowSourceUser)
	}
	if f.Metadata.Version != "1.0.0" {
		t.Errorf("Metadata.Version = %q, want %q", f.Metadata.Version, "1.0.0")
	}
	if f.Raw.Script != "echo hello" {
		t.Errorf("Raw.Script = %q, want %q", f.Raw.Script, "echo hello")
	}
}
