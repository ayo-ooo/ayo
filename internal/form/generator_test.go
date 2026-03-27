package form

import (
	"testing"

	"github.com/ayo-ooo/ayo/internal/schema"
)

func TestGetOrderedProperties_WithInputOrder(t *testing.T) {
	cfg := &FormConfig{
		Schema: &schema.ParsedSchema{
			Properties: map[string]schema.Property{
				"c": {Type: "string"},
				"a": {Type: "string"},
				"b": {Type: "string"},
			},
		},
		InputOrder: []string{"a", "b", "c"},
	}

	g := NewGenerator(cfg)
	order := g.GetOrderedProperties()

	if len(order) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(order))
	}

	expected := []string{"a", "b", "c"}
	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("order[%d] = %q, want %q", i, order[i], exp)
		}
	}
}

func TestBuildField_Required(t *testing.T) {
	cfg := &FormConfig{
		Schema: &schema.ParsedSchema{
			Required: []string{"prompt"},
			Properties: map[string]schema.Property{
				"prompt": {Type: "string", Description: "The prompt"},
			},
		},
	}

	g := NewGenerator(cfg)
	field := g.BuildField("prompt", cfg.Schema.Properties["prompt"])

	if !field.Required {
		t.Error("field.Required should be true")
	}
}

func TestBuildField_Prefill(t *testing.T) {
	cfg := &FormConfig{
		Schema: &schema.ParsedSchema{
			Properties: map[string]schema.Property{
				"name": {Type: "string", Description: "Name"},
			},
		},
		Prefill: map[string]any{
			"name": "test-value",
		},
	}

	g := NewGenerator(cfg)
	field := g.BuildField("name", cfg.Schema.Properties["name"])

	if field.PrefillValue != "test-value" {
		t.Errorf("PrefillValue = %v, want test-value", field.PrefillValue)
	}
}

func TestBuildField_DefaultValue(t *testing.T) {
	cfg := &FormConfig{
		Schema: &schema.ParsedSchema{
			Properties: map[string]schema.Property{
				"scope": {Type: "string", Default: "project"},
			},
		},
	}

	g := NewGenerator(cfg)
	field := g.BuildField("scope", cfg.Schema.Properties["scope"])

	if field.DefaultValue != "project" {
		t.Errorf("DefaultValue = %v, want project", field.DefaultValue)
	}
}

func TestIsInteractiveSupported(t *testing.T) {
	tests := []struct {
		name     string
		prop     schema.Property
		expected bool
	}{
		{"string", schema.Property{Type: "string"}, true},
		{"integer", schema.Property{Type: "integer"}, true},
		{"number", schema.Property{Type: "number"}, true},
		{"boolean", schema.Property{Type: "boolean"}, true},
		{"array with enum", schema.Property{Type: "array", Enum: []string{"a", "b"}}, true},
		{"array without enum", schema.Property{Type: "array"}, false},
		{"object", schema.Property{Type: "object"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInteractiveSupported(tt.prop)
			if result != tt.expected {
				t.Errorf("IsInteractiveSupported() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetMissingRequiredFields(t *testing.T) {
	cfg := &FormConfig{
		Schema: &schema.ParsedSchema{
			Required: []string{"prompt"},
			Properties: map[string]schema.Property{
				"prompt": {Type: "string", Description: "The prompt"},
				"scope":  {Type: "string", Description: "The scope", Default: "project"},
			},
		},
	}

	// Nothing provided
	missing := GetMissingRequiredFields(cfg, nil)
	if len(missing) != 1 || missing[0] != "prompt" {
		t.Errorf("missing = %v, want [prompt]", missing)
	}

	// Partially provided
	missing = GetMissingRequiredFields(cfg, map[string]any{"prompt": "hello"})
	if len(missing) != 0 {
		t.Errorf("missing = %v, want []", missing)
	}
}

func TestFormatMissingFieldsError(t *testing.T) {
	s := &schema.ParsedSchema{
		Properties: map[string]schema.Property{
			"prompt": {Type: "string", Description: "The prompt"},
			"scope":  {Type: "string", Description: "The scope"},
		},
	}

	missing := []string{"prompt", "scope"}
	err := FormatMissingFieldsError(s, missing)

	if len(err) == 0 {
		t.Error("error message should not be empty")
	}
}
