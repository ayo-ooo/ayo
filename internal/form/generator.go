package form

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ayo-ooo/ayo/internal/schema"
	"charm.land/huh/v2"
)

// Generate creates a huh.Form from the schema configuration.
func (g *Generator) Generate() (*huh.Form, error) {
	var fields []huh.Field

	order := g.GetOrderedProperties()

	inOrder := make(map[string]bool)
	for _, name := range order {
		inOrder[name] = true
	}

	for _, name := range order {
		prop, ok := g.config.Schema.Properties[name]
		if !ok {
			continue
		}

		if !IsInteractiveSupported(prop) {
			continue
		}

		field := g.BuildField(name, prop)
		huhField, err := g.fieldToHuh(field)
		if err != nil {
			return nil, fmt.Errorf("converting field %s: %w", name, err)
		}
		fields = append(fields, huhField)
	}

	for name, prop := range g.config.Schema.Properties {
		if inOrder[name] {
			continue
		}

		if !IsInteractiveSupported(prop) {
			continue
		}

		field := g.BuildField(name, prop)
		huhField, err := g.fieldToHuh(field)
		if err != nil {
			return nil, fmt.Errorf("converting field %s: %w", name, err)
		}
		fields = append(fields, huhField)
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no interactive fields available")
	}

	group := huh.NewGroup(fields...)
	form := huh.NewForm(group)

	return form, nil
}

func (g *Generator) fieldToHuh(field FormField) (huh.Field, error) {
	switch {
	case len(field.Options) > 0 && field.Type == "string":
		return g.createSelectField(field)
	case len(field.Options) > 0 && field.Type == "array":
		return g.createMultiSelectField(field)
	case field.Type == "boolean":
		return g.createConfirmField(field)
	case field.Type == "integer" || field.Type == "number":
		return g.createNumericField(field)
	case field.Type == "string":
		return g.createInputField(field)
	default:
		return nil, fmt.Errorf("unsupported field type: %s", field.Type)
	}
}

func (g *Generator) createInputField(field FormField) (huh.Field, error) {
	input := huh.NewInput().
		Title(field.Title).
		Key(field.Name)

	var valueStr string
	if field.PrefillValue != nil {
		if str, ok := field.PrefillValue.(string); ok && str != "" {
			valueStr = str
		}
	} else if field.DefaultValue != nil {
		if str, ok := field.DefaultValue.(string); ok && str != "" {
			valueStr = str
		}
	}
	if valueStr != "" {
		input.Value(&valueStr)
	}

	if field.Required {
		input.Validate(huh.ValidateNotEmpty())
	}

	return input, nil
}

func (g *Generator) createSelectField(field FormField) (huh.Field, error) {
	var options []huh.Option[string]
	for _, opt := range field.Options {
		options = append(options, huh.NewOption(opt, opt))
	}

	selectField := huh.NewSelect[string]().
		Title(field.Title).
		Key(field.Name).
		Options(options...)

	var selected string
	if field.PrefillValue != nil {
		if str, ok := field.PrefillValue.(string); ok {
			selected = str
		}
	} else if field.DefaultValue != nil {
		if str, ok := field.DefaultValue.(string); ok {
			selected = str
		}
	}
	if selected != "" {
		selectField.Value(&selected)
	}

	return selectField, nil
}

func (g *Generator) createMultiSelectField(field FormField) (huh.Field, error) {
	var options []huh.Option[string]
	for _, opt := range field.Options {
		options = append(options, huh.NewOption(opt, opt))
	}

	multiSelect := huh.NewMultiSelect[string]().
		Title(field.Title).
		Key(field.Name).
		Options(options...)

	var selected []string
	if field.PrefillValue != nil {
		if arr, ok := field.PrefillValue.([]string); ok {
			selected = arr
		} else if arr, ok := field.PrefillValue.([]any); ok {
			for _, v := range arr {
				if str, ok := v.(string); ok {
					selected = append(selected, str)
				}
			}
		}
	} else if field.DefaultValue != nil {
		if arr, ok := field.DefaultValue.([]string); ok {
			selected = arr
		} else if arr, ok := field.DefaultValue.([]any); ok {
			for _, v := range arr {
				if str, ok := v.(string); ok {
					selected = append(selected, str)
				}
			}
		}
	}
	if len(selected) > 0 {
		multiSelect.Value(&selected)
	}

	return multiSelect, nil
}

func (g *Generator) createConfirmField(field FormField) (huh.Field, error) {
	confirm := huh.NewConfirm().
		Title(field.Title).
		Key(field.Name)

	var selected bool
	if field.PrefillValue != nil {
		if b, ok := field.PrefillValue.(bool); ok {
			selected = b
		}
	} else if field.DefaultValue != nil {
		if b, ok := field.DefaultValue.(bool); ok {
			selected = b
		}
	}
	confirm.Value(&selected)

	return confirm, nil
}

func (g *Generator) createNumericField(field FormField) (huh.Field, error) {
	input := huh.NewInput().
		Title(field.Title).
		Key(field.Name)

	var valueStr string
	if field.PrefillValue != nil {
		switch v := field.PrefillValue.(type) {
		case int:
			valueStr = strconv.Itoa(v)
		case float64:
			valueStr = strconv.FormatFloat(v, 'f', -1, 64)
		case string:
			valueStr = v
		}
	} else if field.DefaultValue != nil {
		switch v := field.DefaultValue.(type) {
		case int:
			valueStr = strconv.Itoa(v)
		case float64:
			valueStr = strconv.FormatFloat(v, 'f', -1, 64)
		}
	}
	if valueStr != "" {
		input.Value(&valueStr)
	}

	input.Validate(func(s string) error {
		if s == "" && !field.Required {
			return nil
		}
		if field.Type == "integer" {
			_, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("must be an integer")
			}
		} else {
			_, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
		}
		return nil
	})

	return input, nil
}

func CollectResults(form *huh.Form, schema *schema.ParsedSchema) map[string]any {
	results := make(map[string]any)

	for name, prop := range schema.Properties {
		switch prop.Type {
		case "boolean":
			if val := form.Get(name); val != nil {
				results[name] = val
			}
		case "string":
			if val := form.GetString(name); val != "" {
				results[name] = val
			}
		case "integer":
			if val := form.GetString(name); val != "" {
				if i, err := strconv.Atoi(val); err == nil {
					results[name] = i
				}
			}
		case "number":
			if val := form.GetString(name); val != "" {
				if f, err := strconv.ParseFloat(val, 64); err == nil {
					results[name] = f
				}
			}
		case "array":
			if val := form.Get(name); val != nil {
				results[name] = val
			}
		}
	}

	return results
}

func GetMissingRequiredFields(config *FormConfig, provided map[string]any) []string {
	var missing []string

	for _, req := range config.Schema.Required {
		prop, ok := config.Schema.Properties[req]
		if !ok || !IsInteractiveSupported(prop) {
			continue
		}

		val, hasValue := provided[req]
		if !hasValue || val == nil || val == "" {
			if prop.Default == nil {
				missing = append(missing, req)
			}
		}
	}

	return missing
}

func FormatMissingFieldsError(schema *schema.ParsedSchema, missing []string) string {
	var b strings.Builder
	b.WriteString("Missing required arguments:\n\n")

	for _, name := range missing {
		prop := schema.Properties[name]
		desc := prop.Description
		if desc == "" {
			desc = name
		}
		b.WriteString(fmt.Sprintf("  --%-15s %s\n", name, desc))
	}

	b.WriteString("\nRun with --help for usage information.")
	return b.String()
}
