package form

import (
	"github.com/ayo-ooo/ayo/internal/schema"
)

// FormConfig holds the configuration for generating a form from a schema.
type FormConfig struct {
	Schema     *schema.ParsedSchema
	InputOrder []string // Field order from agent config
	Defaults   map[string]any
	Prefill    map[string]any // Values from CLI args
}

// FormField represents a single field in the generated form.
type FormField struct {
	Name         string
	Type         string
	Title        string
	Description  string
	Required     bool
	DefaultValue any
	Options      []string
	PrefillValue any
}

// Generator creates huh forms from JSON Schema.
type Generator struct {
	config *FormConfig
}

// NewGenerator creates a new form generator.
func NewGenerator(config *FormConfig) *Generator {
	return &Generator{config: config}
}

// GetOrderedProperties returns property names in the correct order.
func (g *Generator) GetOrderedProperties() []string {
	if len(g.config.InputOrder) > 0 {
		return g.config.InputOrder
	}

	order := make([]string, 0, len(g.config.Schema.Properties))
	for name := range g.config.Schema.Properties {
		order = append(order, name)
	}
	return order
}

// BuildField creates a FormField from a schema property.
func (g *Generator) BuildField(name string, prop schema.Property) FormField {
	field := FormField{
		Name:         name,
		Type:         prop.Type,
		Title:        prop.Description,
		Description:  prop.Description,
		DefaultValue: prop.Default,
		Options:      prop.Enum,
	}

	if prop.Description != "" {
		field.Title = prop.Description
	} else {
		field.Title = name
	}

	for _, req := range g.config.Schema.Required {
		if req == name {
			field.Required = true
			break
		}
	}

	if val, ok := g.config.Prefill[name]; ok {
		field.PrefillValue = val
	}

	return field
}

// IsInteractiveSupported returns true if the field type can be rendered in a form.
func IsInteractiveSupported(prop schema.Property) bool {
	switch prop.Type {
	case "string", "integer", "number", "boolean":
		return true
	case "array":
		return len(prop.Enum) > 0
	default:
		return false
	}
}
